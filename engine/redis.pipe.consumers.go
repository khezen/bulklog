package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/khezen/bulklog/consumer"
)

func getRedisConsumers(tx redis.Pipeliner, pipeKey string) (consumers map[string]consumer.Interface, err error) {
	consumerStrings, err := tx.HGet(pipeKey, "consumers").Result()
	if err != nil {
		return nil, err
	}
	consumersBytes, err := base64.StdEncoding.DecodeString(consumerStrings)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(consumersBytes)
	err = gob.NewDecoder(buf).Decode(&consumers)
	if err != nil {
		return nil, err
	}
	return consumers, nil
}

func addRedisConsumers(tx redis.Pipeliner, pipeKey string, consumers map[string]consumer.Interface) (err error) {
	key := fmt.Sprintf("%s.consumers", pipeKey)
	var consumerName string
	_, err = tx.Del(key).Result()
	if err != nil {
		return err
	}
	for consumerName = range consumers {
		_, err = tx.RPushX(key, consumerName).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func delRedisConsumer(tx redis.Pipeliner, pipeKey, consumerName string) (err error) {
	_, err = tx.LRem(fmt.Sprintf("%s.consumers", pipeKey), 0, consumerName).Result()
	return err
}
