package engine

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/khezen/bulklog/redisc"
)

var (
	errRedisPipeNotFound = errors.New("errRedisPipeNotFound")
)

func getRedisPipe(red redisc.Connector, pipeKey string) (
	startedAt time.Time,
	retryPeriod, retentionPeriod time.Duration,
	err error) {
	conn, err := red.Open()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("redis.Open.%s", err)
	}
	defer conn.Close()
	err = conn.Send("MULTI")
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("MULTI.%s", err)
	}
	err = conn.Send("HGET", pipeKey, "retryPeriodNano")
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey retryPeriodNano).%s", err)
	}
	err = conn.Send("HGET", pipeKey, "retentionPeriodNano")
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey retentionPeriodNano).%s", err)
	}
	err = conn.Send("HGET", pipeKey, "startedAt")
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey startedAt).%s", err)
	}
	err = conn.Send("EXEC")
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("EXEC.%s", err)
	}
	err = conn.Flush()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("redisConFlush.%s", err)
	}
	retryPeriodStr, err := conn.Receive()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey retryPeriodNano).%s", err)
	}
	if retryPeriodStr == nil {
		return time.Time{}, 0, 0, errRedisPipeNotFound
	}
	retryPeriodInt, err := strconv.Atoi(retryPeriodStr.(string))
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("retryPeriodAtoi.%s", err)
	}
	retryPeriod = time.Duration(retryPeriodInt)
	retentionPeriodStr, err := conn.Receive()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey retentionPeriodNano).%s", err)
	}
	retentionPeriodInt, err := strconv.Atoi(retentionPeriodStr.(string))
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("retentionPeriodAtoi.%s", err)
	}
	retentionPeriod = time.Duration(retentionPeriodInt)
	startedAtStr, err := conn.Receive()
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("(HGET pipeKey startedAt).%s", err)
	}
	startedAt, err = time.Parse(time.RFC3339Nano, startedAtStr.(string))
	if err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("parseStartedAtStr.%s", err)
	}
	return startedAt, retryPeriod, retentionPeriod, nil
}

func newRedisPipe(conn redis.Conn, pipeKey string, retryPeriod, retentionPeriod time.Duration, startedAt time.Time) (err error) {
	err = conn.Send("HSET", pipeKey, "retryPeriodNano", int64(retryPeriod))
	if err != nil {
		return fmt.Errorf("(HSET pipeKey retryPeriodNano %d).%s", retryPeriod, err)
	}
	err = conn.Send("HSET", pipeKey, "retentionPeriodNano", int64(retentionPeriod))
	if err != nil {
		return fmt.Errorf("(HSET pipeKey retentionPeriodNano %d).%s", retentionPeriod, err)
	}
	err = conn.Send("HSET", pipeKey, "startedAt", startedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("(HSET pipeKey startedAt %s).%s", startedAt.Format(time.RFC3339Nano), err)
	}
	err = setRedisPipeIteration(conn, pipeKey, 0)
	if err != nil {
		return fmt.Errorf("setRedisPipeIteration.%s", err)
	}
	return nil
}

func deleteRedisPipe(red redisc.Connector, pipeKey string) (err error) {
	conn, err := red.Open()
	if err != nil {
		return fmt.Errorf("redis.Open.%s", err)
	}
	defer conn.Close()
	err = conn.Send("DEL", pipeKey)
	if err != nil {
		return fmt.Errorf("DEL pipeKey).%s", err)
	}
	err = deleteRedisPipeConsumers(conn, pipeKey)
	if err != nil {
		return fmt.Errorf("deleteRedisPipeConsumers.%s", err)
	}
	err = deleteRedisPipeDocuments(conn, pipeKey)
	if err != nil {
		return fmt.Errorf("deleteRedisPipeDocuments.%s", err)
	}
	return nil
}
