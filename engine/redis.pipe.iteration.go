package engine

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

func getRedisPipeIteration(tx redis.Pipeliner, pipeKey string) (i int, err error) {
	iStr, err := tx.HGet(pipeKey, "iteration").Result()
	if err != nil {
		return -1, fmt.Errorf("(HGET pipeKey iteration).%s", err.Error())
	}
	return strconv.Atoi(iStr)
}

func setRedisPipeIteration(tx redis.Pipeliner, pipeKey string, iter int) (err error) {
	_, err = tx.HSet(pipeKey, "iteration", iter).Result()
	if err != nil {
		return fmt.Errorf("(HSET pipeKey iteration %d).%s", iter, err.Error())
	}
	return nil
}

func incrRedisPipeIteration(tx redis.Pipeliner, pipeKey string) (err error) {
	_, err = tx.HIncrBy(pipeKey, "iteration", 1).Result()
	if err != nil {
		return fmt.Errorf("(HINCRBY pipeKey iteration 1).%s", err.Error())
	}
	return nil
}
