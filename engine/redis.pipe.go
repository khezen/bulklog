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
	var stringCmder *redis.StringCmd
	stringCmder = red.HGet(pipeKey, "retryPeriodNano")
	err = stringCmder.Err()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey retryPeriodNano).%s", err.Error())
	}
	retryPeriodStr := stringCmder.Val()
	retryPeriodInt, err := strconv.Atoi(retryPeriodStr)
	retryPeriod = time.Duration(retryPeriodInt)
	stringCmder = red.HGet(pipeKey, "retentionPeriodNano")
	err = stringCmder.Err()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey retentionPeriodNano).%s", err.Error())
	}
	retentionPeriodStr := stringCmder.Val()
	retentionPeriodInt, err := strconv.Atoi(retentionPeriodStr)
	retentionPeriod = time.Duration(retentionPeriodInt)
	stringCmder = red.HGet(pipeKey, "startedAt")
	err = stringCmder.Err()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey startedAt).%s", err.Error())
	}
	startedAtStr := stringCmder.Val()
	startedAt, err = time.Parse(time.RFC3339Nano, startedAtStr)
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("parseStartedAtStr.%s", err.Error())
	}
	return startedAt, retryPeriod, retentionPeriod, nil
}

func newRedisPipe(tx *redis.Tx, pipeKey string, retryPeriod, retentionPeriod time.Duration, startedAt time.Time) (err error) {
	var boolCmder *redis.BoolCmd
	boolCmder = tx.HSet(pipeKey, "retryPeriodNano", int64(retryPeriod))
	err = boolCmder.Err()
	if err != nil {
		return fmt.Errorf("(HSET pipeKey retryPeriodNano %d).%s", retryPeriod, err.Error())
	}
	boolCmder = tx.HSet(pipeKey, "retentionPeriodNano", int64(retentionPeriod))
	err = boolCmder.Err()
	if err != nil {
		return fmt.Errorf("(HSET pipeKey retentionPeriodNano %d).%s", retentionPeriod, err.Error())
	}
	boolCmder = tx.HSet(pipeKey, "startedAt", startedAt.Format(time.RFC3339Nano))
	err = boolCmder.Err()
	if err != nil {
		return fmt.Errorf("(HSET pipeKey startedAt %s).%s", startedAt.Format(time.RFC3339Nano), err.Error())
	}
	err = setRedisPipeIteration(tx, pipeKey, 0)
	if err != nil {
		return fmt.Errorf("setRedisPipeIteration.%s", err.Error())
	}
	return nil
}

func deleteRedisPipe(tx *redis.Tx, pipeKey string) (err error) {
	var intCmder *redis.IntCmd
	intCmder = tx.Del(pipeKey)
	err = intCmder.Err()
	if err != nil {
		return fmt.Errorf("DEL pipeKey).%s", err)
	}
	err = deleteRedisPipeConsumers(tx, pipeKey)
	if err != nil {
		return fmt.Errorf("deleteRedisPipeConsumers.%s", err.Error())
	}
	err = deleteRedisPipeDocuments(tx, pipeKey)
	if err != nil {
		return fmt.Errorf("deleteRedisPipeDocuments.%s", err.Error())
	}
	return nil
}
