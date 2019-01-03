package engine

import (
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

func getRedisPipeIteration(red redis.Conn, pipeKey string) (i int, err error) {
	iStr, err := red.Do("HGET", pipeKey, "iteration")
	if err != nil {
		return -1, fmt.Errorf("(HGET pipeKey iteration).%s", err)
	}
	if iStr == nil {
		return 0, nil
	}
	return strconv.Atoi(iStr.(string))
}

func setRedisPipeIteration(red redis.Conn, pipeKey string, iter int) (err error) {
	err = red.Send("HSET", pipeKey, "iteration", iter)
	if err != nil {
		return fmt.Errorf("(HSET pipeKey iteration %d).%s", iter, err)
	}
	return nil
}

func incrRedisPipeIteration(red redis.Conn, pipeKey string) (err error) {
	_, err = red.Do("HINCRBY", pipeKey, "iteration", 1)
	if err != nil {
		return fmt.Errorf("(HINCRBY pipeKey iteration 1).%s", err)
	}
	return nil
}
