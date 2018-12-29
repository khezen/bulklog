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
		return nil, fmt.Errorf("(LLEN pipeKey.consumers).%s", err)
	}
	remainingConsumerNames, err := tx.LRange(key, 0, remainingConsumersLen).Result()
	if err != nil {
		return nil, fmt.Errorf("(LRANGE pipeKey.consumers).%s", err)
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
	consumerNames := make([]interface{}, 0, len(consumers))
	var consumerName string
	for consumerName = range consumers {
		consumerNames = append(consumerNames, consumerName)
	}
	_, err = tx.RPush(key, consumerNames...).Result()
	if err != nil {
		return fmt.Errorf("(RPUSH pipeKey.consumers consumerNames...).%s", err)
	}
	return
}

func deleteRedisPipeConsumer(tx redis.Pipeliner, pipeKey, consumerName string) (err error) {
	_, err = tx.LRem(fmt.Sprintf("%s.consumers", pipeKey), 0, consumerName).Result()
	if err != nil {
		return fmt.Errorf("(LREM pipeKey.consumers consumerName).%s", err)
	}
	return nil
}

func deleteRedisPipeConsumers(tx redis.Pipeliner, pipeKey string) (err error) {
	_, err = tx.Del(fmt.Sprintf("%s.consumers", pipeKey)).Result()
	if err != nil {
		return fmt.Errorf("(DEL pipeKey.consumers).%s", err)
	}
	return nil
}
