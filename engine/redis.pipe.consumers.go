package engine

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/khezen/bulklog/consumer"
)

func getRedisPipeConsumers(tx redis.Pipeliner, pipeKey string, consumers map[string]consumer.Interface) (remainingConsumers map[string]consumer.Interface, err error) {
	key := fmt.Sprintf("%s.consumers", pipeKey)
	remainingConsumersLen, err := tx.LLen(key).Result()
	if err != nil {
		return nil, err
	}
	remainingConsumerNames, err := tx.LRange(key, 0, remainingConsumersLen).Result()
	if err != nil {
		return nil, err
	}
	remainingConsumers = make(map[string]consumer.Interface)
	for _, consumerName := range remainingConsumerNames {
		if cons, ok := consumers[consumerName]; ok {
			remainingConsumers[consumerName] = cons
		}
	}
	return remainingConsumers, nil
}

func addRedisPipeConsumers(tx redis.Pipeliner, pipeKey string, consumers map[string]consumer.Interface) (err error) {
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

func deleteRedisPipeConsumer(tx redis.Pipeliner, pipeKey, consumerName string) (err error) {
	_, err = tx.LRem(fmt.Sprintf("%s.consumers", pipeKey), 0, consumerName).Result()
	return err
}

func deleteRedisPipeConsumers(tx redis.Pipeliner, pipeKey string) (err error) {
	_, err = tx.Del(fmt.Sprintf("%s.consumers", pipeKey)).Result()
	return err
}
