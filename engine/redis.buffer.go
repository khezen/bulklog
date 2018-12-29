package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/khezen/bulklog/collection"
	"github.com/khezen/bulklog/config"
	"github.com/khezen/bulklog/consumer"
)

type redisBuffer struct {
	redis         *redis.Client
	collection    *collection.Collection
	consumers     map[string]consumer.Interface
	bufferKey     string
	timeKey       string
	pipeKeyPrefix string
	flushedAt     time.Time
	close         chan struct{}
}

// RedisBuffer -
func RedisBuffer(collec *collection.Collection, redisConfig config.Redis, consumers map[string]consumer.Interface) Buffer {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Endpoint,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	rbuffer := &redisBuffer{
		redis:         redisClient,
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
	_, err = b.redis.RPush(b.bufferKey, docBase64).Result()
	if err != nil {
		return fmt.Errorf("(RPUSH collection.buffer docBase64).%s", err)
	}
	return nil
}

func (b *redisBuffer) Flush() (err error) {
	now := time.Now().UTC()
	tx := b.redis.TxPipeline()
	defer func() {
		if err != nil {
			err = tx.Discard()
			if err != nil {
				fmt.Printf("DISCARD.%s)\n", err)
			}
		} else {
			cmders, err := tx.Exec()
			if err != nil {
				fmt.Printf("EXEC.%v.%s)\n", cmders, err)
			}
		}
	}()
	flushedAtStr, err := tx.Get(b.timeKey).Result()
	if err != nil {
		return fmt.Errorf("(GET collection.flushedAt).%s", err)
	}
	if flushedAtStr != "" {
		b.flushedAt, err = time.Parse(time.RFC3339Nano, flushedAtStr)
		if err != nil {
			return fmt.Errorf("parseFlushedAtStr.%s", err)
		}
	}
	if time.Since(b.flushedAt) < b.collection.FlushPeriod {
		return
	}
	var length int64
	length, err = tx.LLen(b.bufferKey).Result()
	if err != nil {
		return fmt.Errorf("(LLEN bufferKey).%s", err.Error())
	}
	if length == 0 {
		return
	}
	pipeID := uuid.New()
	pipeKey := fmt.Sprintf("%s.%s", b.pipeKeyPrefix, pipeID)
	err = newRedisPipe(tx, pipeKey, b.collection.FlushPeriod, b.collection.RetentionPeriod, now)
	if err != nil {
		return fmt.Errorf("newRedisPipe.%s", err)
	}
	err = addRedisPipeConsumers(tx, pipeKey, b.consumers)
	if err != nil {
		return fmt.Errorf("addRedisPipeConsumers.%s", err)
	}
	err = flushBuffer2RedisPipe(tx, b.bufferKey, pipeKey)
	if err != nil {
		return fmt.Errorf("flushBuffer2RedisPipe.%s", err)
	}
	_, err = tx.Set(b.timeKey, now.Format(time.RFC3339Nano), 0).Result()
	if err != nil {
		return fmt.Errorf("(SET collection.flushedAt %s).%s", now.Format(time.RFC3339Nano), err)
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
