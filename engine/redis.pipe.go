package engine

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func getPipe(red redis.Client, pipeKey string) (
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

func deletePipe(tx redis.Pipeliner, pipeKey string) (err error) {
	_, err = tx.Del(pipeKey).Result()
	if err != nil {
		return err
	}
	_, err = tx.Del(fmt.Sprintf("%s.buffer", pipeKey)).Result()
	if err != nil {
		return err
	}
	return nil
}
