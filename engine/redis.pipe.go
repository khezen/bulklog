package engine

import (
	"fmt"
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
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey retryPeriodNano).%s", err.Error())
	}
	retryPeriodInt, err := strconv.Atoi(retryPeriodStr)
	retryPeriod = time.Duration(retryPeriodInt)
	retentionPeriodStr, err := red.HGet(pipeKey, "retentionPeriodNano").Result()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey retentionPeriodNano).%s", err.Error())
	}
	retentionPeriodInt, err := strconv.Atoi(retentionPeriodStr)
	retentionPeriod = time.Duration(retentionPeriodInt)
	startedAtStr, err := red.HGet(pipeKey, "startedAt").Result()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey startedAt).%s", err.Error())
	}
	startedAt, err = time.Parse(time.RFC3339Nano, startedAtStr)
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("parseStartedAtStr.%s", err.Error())
	}
	return startedAt, retryPeriod, retentionPeriod, nil
}

func newRedisPipe(tx redis.Pipeliner, pipeKey string, retryPeriod, retentionPeriod time.Duration, startedAt time.Time) (err error) {
	_, err = tx.HSet(pipeKey, "retryPeriodNano", int64(retryPeriod)).Result()
	if err != nil {
		return fmt.Errorf("(HSET pipeKey retryPeriodNano %d).%s", retryPeriod, err.Error())
	}
	_, err = tx.HSet(pipeKey, "retentionPeriodNano", int64(retentionPeriod)).Result()
	if err != nil {
		return fmt.Errorf("(HSET pipeKey retentionPeriodNano %d).%s", retentionPeriod, err.Error())
	}
	_, err = tx.HSet(pipeKey, "startedAt", startedAt.Format(time.RFC3339Nano)).Result()
	if err != nil {
		return fmt.Errorf("(HSET pipeKey startedAt %s).%s", startedAt.Format(time.RFC3339Nano), err.Error())
	}
	err = setRedisPipeIteration(tx, pipeKey, 0)
	if err != nil {
		return fmt.Errorf("setRedisPipeIteration.%s", err.Error())
	}
	return nil
}

func deleteRedisPipe(tx redis.Pipeliner, pipeKey string) (err error) {
	_, err = tx.Del(pipeKey).Result()
	if err != nil {
		return fmt.Errorf("DEL pipeKey).%s", err)
	}
	err = deleteRedisPipeConsumers(tx, pipeKey)
	if err != nil {
		return fmt.Errorf("deleteRedisPipeConsumers.%s", err.Error())
	}
	err = deleteRedisPipeDocuments(tx, pipeKey)
	return fmt.Errorf("deleteRedisPipeDocuments.%s", err.Error())
}
