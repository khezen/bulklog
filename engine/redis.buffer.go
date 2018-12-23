package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/khezen/espipe/collection"
	"github.com/khezen/espipe/config"
	"github.com/khezen/espipe/consumer"
)

type redisBuffer struct {
	sync.Mutex
	redis         redis.Client
	collection    collection.Collection
	consumers     []consumer.Interface
	bufferKey     string
	timeKey       string
	pipeKeyPrefix string
	flushedAt     time.Time
	close         chan struct{}
}

// RedisBuffer -
func RedisBuffer(collec collection.Collection, redisConfig config.Redis, consumers ...consumer.Interface) (Buffer, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Address,
		Password: redisConfig.Password,
		DB:       redisConfig.Partition,
	})
	rbuffer := &redisBuffer{
		Mutex:         sync.Mutex{},
		redis:         *redisClient,
		collection:    collec,
		consumers:     consumers,
		bufferKey:     fmt.Sprintf("espipe.%s.buffer", collec.Name),
		timeKey:       fmt.Sprintf("espipe.%s.flushedAt", collec.Name),
		pipeKeyPrefix: fmt.Sprintf("espipe.%s.pipes", collec.Name),
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

func (b *redisBuffer) Set(consumers ...consumer.Interface) {
	b.Lock()
	b.consumers = consumers
	b.Unlock()
}

func (b *redisBuffer) Append(doc collection.Document) (err error) {
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(doc)
	if err != nil {
		return err
	}
	docBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	_, err = b.redis.RPushX(b.bufferKey, docBase64).Result()
	if err != nil {
		return err
	}
	return nil
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
	latestFlushAt, err := time.Parse(time.RFC3339Nano, latestFlushAtStr)
	if err != nil {
		return err
	}
	if time.Since(latestFlushAt) < b.collection.FlushPeriod {
		b.flushedAt = latestFlushAt
		return nil
	}
	pipeID := uuid.New()
	pipeKey := fmt.Sprintf("%s.%s", b.pipeKeyPrefix, pipeID)
	_, err = tx.Rename(b.bufferKey, fmt.Sprintf("%s.buffer", pipeKey)).Result()
	if err != nil {
		return err
	}
	_, err = tx.HSet(pipeKey, "retryPeriodNano", b.collection.FlushPeriod).Result()
	if err != nil {
		return err
	}
	_, err = tx.HSet(pipeKey, "retentionPeriodNano", b.collection.RetentionPeriod).Result()
	if err != nil {
		return err
	}
	_, err = tx.HSet(pipeKey, "startedAt", now.Format(time.RFC3339Nano)).Result()
	if err != nil {
		return err
	}
	_, err = tx.HSet(pipeKey, "iter", 0).Result()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(b.consumers)
	if err != nil {
		return err
	}
	consumersBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	_, err = tx.HSet(pipeKey, "consumers", consumersBase64).Result()
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
