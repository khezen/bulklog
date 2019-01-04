package engine

import (
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

func getRedisPipeIteration(red *redis.Pool, pipeKey string) (i int, err error) {
	conn := red.Get()
	defer conn.Close()
	iStr, err := conn.Do("HGET", pipeKey, "iteration")
	if err != nil {
		return -1, fmt.Errorf("(HGET pipeKey iteration).%s", err)
	}
	if iStr == nil {
		return 0, nil
	}
	return strconv.Atoi(string(iStr.([]byte)))
}

func setRedisPipeIteration(conn redis.Conn, pipeKey string, iter int) (err error) {
	err = conn.Send("HSET", pipeKey, "iteration", iter)
	if err != nil {
		return fmt.Errorf("(HSET pipeKey iteration %d).%s", iter, err)
	}
	return nil
}

func incrRedisPipeIteration(red *redis.Pool, pipeKey string) (err error) {
	conn := red.Get()
	defer conn.Close()
	_, err = conn.Do("HINCRBY", pipeKey, "iteration", 1)
	if err != nil {
		return fmt.Errorf("(HINCRBY pipeKey iteration 1).%s", err)
	}
	return nil
}
