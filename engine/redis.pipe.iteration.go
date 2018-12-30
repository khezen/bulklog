package engine

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

func getRedisPipeIteration(tx *redis.Tx, pipeKey string) (i int, err error) {
	var cmder *redis.StringCmd
	cmder = tx.HGet(pipeKey, "iteration")
	err = cmder.Err()
	if err != nil {
		return -1, fmt.Errorf("(HGET pipeKey iteration).%s", err)
	}
	return strconv.Atoi(cmder.Val())
}

func setRedisPipeIteration(tx *redis.Tx, pipeKey string, iter int) (err error) {
	var cmder *redis.BoolCmd
	cmder = tx.HSet(pipeKey, "iteration", iter)
	err = cmder.Err()
	if err != nil {
		return fmt.Errorf("(HSET pipeKey iteration %d).%s", iter, err)
	}
	return nil
}

func incrRedisPipeIteration(tx *redis.Tx, pipeKey string) (err error) {
	var cmder *redis.IntCmd
	cmder = tx.HIncrBy(pipeKey, "iteration", 1)
	err = cmder.Err()
	if err != nil {
		return fmt.Errorf("(HINCRBY pipeKey iteration 1).%s", err)
	}
	return nil
}
