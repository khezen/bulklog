package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/khezen/bulklog/collection"
	"github.com/khezen/bulklog/config"
	"github.com/khezen/bulklog/consumer"
)

type redisBuffer struct {
	redisCon      redis.Conn
	collection    *collection.Collection
	consumers     map[string]consumer.Interface
	bufferKey     string
	timeKey       string
	pipeKeyPrefix string
	flushedAt     time.Time
	close         chan struct{}
}

// RedisBuffer -
func RedisBuffer(collec *collection.Collection, redisConfig config.Redis, consumers map[string]consumer.Interface) (Buffer, error) {
	redisCon, err := redis.Dial(
		"tcp", redisConfig.Endpoint,
		redis.DialDatabase(redisConfig.DB),
		redis.DialPassword(redisConfig.Password),
	)
	if err != nil {
		return nil, err
	}
	rbuffer := &redisBuffer{
		redisCon:      redisCon,
		collection:    collec,
		consumers:     consumers,
		bufferKey:     fmt.Sprintf("bulklog.%s.buffer", collec.Name),
		timeKey:       fmt.Sprintf("bulklog.%s.flushedAt", collec.Name),
		pipeKeyPrefix: fmt.Sprintf("bulklog.%s.pipes", collec.Name),
		flushedAt:     time.Now().UTC(),
		close:         make(chan struct{}),
	}
	redisConveyAll(rbuffer.redisCon, rbuffer.pipeKeyPrefix, rbuffer.consumers)
	return rbuffer, nil
}

func (b *redisBuffer) Append(doc *collection.Document) (err error) {
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(*doc)
	if err != nil {
		return
	}
	docBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	_, err = b.redisCon.Do("RPUSH", b.bufferKey, docBase64)
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

	flushedAtStr, err := b.redisCon.Do("GET", b.timeKey)
	if err != nil {
		return fmt.Errorf("(GET collection.flushedAt).%s", err)
	}
	if flushedAtStr != "" {
		b.flushedAt, err = time.Parse(time.RFC3339Nano, flushedAtStr.(string))
		if err != nil {
			return fmt.Errorf("parseFlushedAtStr.%s", err)
		}
	}
	if time.Since(b.flushedAt) < b.collection.FlushPeriod {
		return
	}
	length, err := b.redisCon.Do("LLEN", b.bufferKey)
	if err != nil {
		return fmt.Errorf("(LLEN bufferKey).%s", err)
	}
	if length == 0 {
		_, err = b.redisCon.Do("SET", b.timeKey, now.Format(time.RFC3339Nano), 0)
		if err != nil {
			return fmt.Errorf("(SET collection.flushedAt %s).%s", now.Format(time.RFC3339Nano), err)
		}
		b.flushedAt = now
		return
	}
	pipeID = uuid.New()
	pipeKey = fmt.Sprintf("%s.%s", b.pipeKeyPrefix, pipeID)
	err = b.redisCon.Send("MULTI")
	if err != nil {
		return fmt.Errorf("MULTI.%s", err)
	}
	err = newRedisPipe(b.redisCon, pipeKey, b.collection.FlushPeriod, b.collection.RetentionPeriod, now)
	if err != nil {
		return fmt.Errorf("newRedisPipe.%s", err)
	}
	err = addRedisPipeConsumers(b.redisCon, pipeKey, b.consumers)
	if err != nil {
		return fmt.Errorf("addRedisPipeConsumers.%s", err)
	}
	err = flushBuffer2RedisPipe(b.redisCon, b.bufferKey, pipeKey)
	if err != nil {
		return fmt.Errorf("flushBuffer2RedisPipe.%s", err)
	}
	err = b.redisCon.Send("SET", b.timeKey, now.Format(time.RFC3339Nano), 0)
	if err != nil {
		return fmt.Errorf("(SET collection.flushedAt %s).%s", now.Format(time.RFC3339Nano), err)
	}
	b.flushedAt = now
	_, err = b.redisCon.Do("EXEC")
	if err != nil {
		return fmt.Errorf("EXEC.%s", err)
	}
	go presetRedisConvey(b.redisCon, pipeKey, b.consumers, now, b.collection.FlushPeriod, b.collection.RetentionPeriod)
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
