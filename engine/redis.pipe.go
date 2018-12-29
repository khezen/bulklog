package engine

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func getRedisPipe(red *redis.Client, pipeKey string) (
	startedAt time.Time,
	retryPeriod, retentionPeriod time.Duration,
	err error) {
	retryPeriodStr, err := red.HGet(pipeKey, "retryPeriodNano").Result()
	if err != nil {
		return time.Time{}, 0, 0, err
	}
	retryPeriodInt, err := strconv.Atoi(retryPeriodStr)
	retryPeriod = time.Duration(retryPeriodInt)
	retentionPeriodStr, err := red.HGet(pipeKey, "retentionPeriod").Result()
	if err != nil {
		return time.Time{}, 0, 0, err
	}
	retentionPeriodInt, err := strconv.Atoi(retentionPeriodStr)
	retentionPeriod = time.Duration(retentionPeriodInt)
	startedAtStr, err := red.HGet(pipeKey, "startedAt").Result()
	if err != nil {
		return time.Time{}, 0, 0, err
	}
	startedAt, err = time.Parse(time.RFC3339Nano, startedAtStr)
	if err != nil {
		return time.Time{}, 0, 0, err
	}
	return startedAt, retryPeriod, retentionPeriod, nil
}

func newRedisPipe(tx redis.Pipeliner, pipeKey string, flushPeriod, retentionPeriod time.Duration, startedAt time.Time) (err error) {
	_, err = tx.HSet(pipeKey, "retryPeriodNano", flushPeriod).Result()
	if err != nil {
		return err
	}
	_, err = tx.HSet(pipeKey, "retentionPeriodNano", retentionPeriod).Result()
	if err != nil {
		return err
	}
	_, err = tx.HSet(pipeKey, "startedAt", startedAt.Format(time.RFC3339Nano)).Result()
	if err != nil {
		return err
	}
	err = setRedisPipeIteration(tx, pipeKey, 0)
	return err
}

func deleteRedisPipe(tx redis.Pipeliner, pipeKey string) (err error) {
	_, err = tx.Del(pipeKey).Result()
	if err != nil {
		return err
	}
	err = deleteRedisPipeConsumers(tx, pipeKey)
	if err != nil {
		return err
	}
	err = deleteRedisPipeDocuments(tx, pipeKey)
	return err
}
