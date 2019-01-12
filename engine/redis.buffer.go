package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/bulklog/bulklog/collection"
	"github.com/bulklog/bulklog/config"
	"github.com/bulklog/bulklog/consumer"
)

type redisBuffer struct {
	redis         *redis.Pool
	collection    *collection.Collection
	consumers     map[string]consumer.Interface
	bufferKey     string
	timeKey       string
	pipeKeyPrefix string
	flushedAt     time.Time
	close         chan struct{}
}

// RedisBuffer -
func RedisBuffer(collec *collection.Collection, redisCfg *config.Redis, consumers map[string]consumer.Interface) Buffer {
	rbuffer := &redisBuffer{
		redis: &redis.Pool{
			MaxActive:   redisCfg.MaxConn,
			Wait:        true,
			MaxIdle:     redisCfg.IdleConn,
			IdleTimeout: 5 * time.Minute,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", redisCfg.Endpoint)
				if err != nil {
					return nil, err
				}
				if redisCfg.Password != "" {
					if _, err := c.Do("AUTH", redisCfg.Password); err != nil {
						c.Close()
						return nil, err
					}
				}
				if _, err := c.Do("SELECT", redisCfg.DB); err != nil {
					c.Close()
					return nil, err
				}
				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, err := c.Do("PING")
				return err
			},
		},
		collection:    collec,
		consumers:     consumers,
		bufferKey:     fmt.Sprintf("bulklog.%s.buffer", collec.Name),
		timeKey:       fmt.Sprintf("bulklog.%s.flushedAt", collec.Name),
		pipeKeyPrefix: fmt.Sprintf("bulklog.%s.pipes", collec.Name),
		flushedAt:     time.Now().UTC(),
		close:         make(chan struct{}),
	}
	redisConveyAll(rbuffer.redis, rbuffer.pipeKeyPrefix, rbuffer.consumers)
	return rbuffer
}

func (b *redisBuffer) Append(doc *collection.Document) (err error) {
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(*doc)
	if err != nil {
		return
	}
	docBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	conn := b.redis.Get()
	defer conn.Close()
	_, err = conn.Do("RPUSH", b.bufferKey, docBase64)
	if err != nil {
		return fmt.Errorf("(RPUSH collection.buffer docBase64).%s", err)
	}
	return nil
}

func (b *redisBuffer) Flush() (err error) {
	var (
		now     = time.Now().UTC()
		pipeID  = uuid.New()
		pipeKey = fmt.Sprintf("%s.%s", b.pipeKeyPrefix, pipeID)
	)
	conn := b.redis.Get()
	defer conn.Close()
	flushedAtStr, err := conn.Do("GET", b.timeKey)
	if err != nil {
		return fmt.Errorf("(GET collection.flushedAt).%s", err)
	}
	if flushedAtStr != "" {
		b.flushedAt, err = time.Parse(time.RFC3339Nano, string(flushedAtStr.([]byte)))
		if err != nil {
			return fmt.Errorf("parseFlushedAtStr.%s", err)
		}
	}
	if time.Since(b.flushedAt) < b.collection.FlushPeriod {
		return
	}
	bufferLen, err := conn.Do("LLEN", b.bufferKey)
	if err != nil {
		return fmt.Errorf("(LLEN bufferKey).%s", err)
	}
	if bufferLen == 0 {
		_, err = conn.Do("SET", b.timeKey, now.Format(time.RFC3339Nano), 0)
		if err != nil {
			return fmt.Errorf("(SET collection.flushedAt %s).%s", now.Format(time.RFC3339Nano), err)
		}
		b.flushedAt = now
		return
	}
	pipeID = uuid.New()
	pipeKey = fmt.Sprintf("%s.%s", b.pipeKeyPrefix, pipeID)
	err = conn.Send("MULTI")
	if err != nil {
		return fmt.Errorf("MULTI.%s", err)
	}
	err = newRedisPipe(conn, pipeKey, b.collection.FlushPeriod, b.collection.RetentionPeriod, now)
	if err != nil {
		return fmt.Errorf("newRedisPipe.%s", err)
	}
	err = addRedisPipeConsumers(conn, pipeKey, b.consumers)
	if err != nil {
		return fmt.Errorf("addRedisPipeConsumers.%s", err)
	}
	err = flushBuffer2RedisPipe(conn, b.bufferKey, pipeKey)
	if err != nil {
		return fmt.Errorf("flushBuffer2RedisPipe.%s", err)
	}
	err = conn.Send("SET", b.timeKey, now.Format(time.RFC3339Nano), 0)
	if err != nil {
		return fmt.Errorf("(SET collection.flushedAt %s).%s", now.Format(time.RFC3339Nano), err)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		return fmt.Errorf("EXEC.%s", err)
	}
	b.flushedAt = now
	go presetRedisConvey(b.redis, pipeKey, b.consumers, now, b.collection.FlushPeriod, b.collection.RetentionPeriod)
	return nil
}

// Flusher flushes every tick
func (b *redisBuffer) Flusher() func() {
	return func() {
		var (
			timer   *time.Timer
			waitFor time.Duration
			err     error
		)
		for {
			waitFor = b.collection.FlushPeriod - time.Since(b.flushedAt)
			if waitFor <= 0 {
				err := b.Flush()
				if err != nil {
					fmt.Printf("Flush.%s)\n", err)
					timer = time.NewTimer(time.Second)
					<-timer.C
				}
				continue
			}
			timer = time.NewTimer(waitFor)
			select {
			case <-b.close:
				return
			case <-timer.C:
				err = b.Flush()
				if err != nil {
					fmt.Printf("Flush.%s)\n", err)
				}
				break
			}
		}
	}
}

func (b *redisBuffer) Close() {
	b.close <- struct{}{}
}
