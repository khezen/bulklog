package engine

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

func getIter(red redis.Client, pipeKey string) (i int, err error) {
	iStr, err := red.HGet(pipeKey, "iter").Result()
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(iStr)
}

func setIter(red redis.Client, pipeKey string, iter int) (err error) {
	_, err = red.HSet(pipeKey, "iter", iter).Result()
	if err != nil {
		fmt.Println(err)
		return
	}
	return nil
}
