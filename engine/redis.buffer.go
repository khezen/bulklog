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
func RedisBuffer(collec *collection.Collection, redisConfig config.Redis, consumers map[string]consumer.Interface) (Buffer, error) {
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
	_, err := rbuffer.redis.Set(rbuffer.timeKey, rbuffer.flushedAt.Format(time.RFC3339Nano), 0).Result()
	if err != nil {
		return nil, err
	}
	redisConveyAll(rbuffer.redis, rbuffer.pipeKeyPrefix)
	return rbuffer, nil
}

func (b *redisBuffer) Append(doc *collection.Document) (err error) {
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(*doc)
	if err != nil {
		return
	}
	docBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	_, err = b.redis.RPushX(b.bufferKey, docBase64).Result()
	return
}

func (b *redisBuffer) Flush() (err error) {
	now := time.Now().UTC()
	tx := b.redis.TxPipeline()
	defer func() {
		if err != nil {
			tx.Discard()
		} else {
			tx.Exec()
		}
	}()
	latestFlushAtStr, err := tx.Get(b.timeKey).Result()
	if err != nil {
		return err
	}
	var latestFlushAt time.Time
	if latestFlushAtStr != "" {
		latestFlushAt, err = time.Parse(time.RFC3339Nano, latestFlushAtStr)
		if err != nil {
			return err
		}
	}
	if time.Since(latestFlushAt) < b.collection.FlushPeriod {
		b.flushedAt = latestFlushAt
		return nil
	}
	pipeID := uuid.New()
	pipeKey := fmt.Sprintf("%s.%s", b.pipeKeyPrefix, pipeID)
	err = newRedisPipe(tx, pipeKey, b.collection.FlushPeriod, b.collection.RetentionPeriod, now)
	if err != nil {
		return err
	}
	err = setRedisDocuments(tx, b.bufferKey, pipeKey)
	if err != nil {
		return err
	}
	err = addRedisConsumers(tx, pipeKey, b.consumers)
	if err != nil {
		return err
	}
	_, err = tx.Set(b.timeKey, now.Format(time.RFC3339Nano), 0).Result()
	if err != nil {
		return err
	}
	b.flushedAt = now
	go presetRedisConvey(b.redis, pipeKey, now, b.collection.FlushPeriod, b.collection.RetentionPeriod)
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
					fmt.Println(err)
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
					fmt.Println(err)
				}
				break
			}
		}
	}
}

func (b *redisBuffer) Close() {
	b.close <- struct{}{}
}
