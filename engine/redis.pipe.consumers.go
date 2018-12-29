package engine

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/khezen/bulklog/consumer"
)

func getRedisPipeConsumers(tx *redis.Tx, pipeKey string, consumers map[string]consumer.Interface) (remainingConsumers map[string]consumer.Interface, err error) {
	var (
		key        = fmt.Sprintf("%s.consumers", pipeKey)
		intCmder   *redis.IntCmd
		sliceCmder *redis.StringSliceCmd
	)
	intCmder = tx.LLen(key)
	err = intCmder.Err()
	if err != nil {
		return nil, fmt.Errorf("(LLEN pipeKey.consumers).%s", err)
	}
	remainingConsumersLen := intCmder.Val()
	sliceCmder = tx.LRange(key, 0, remainingConsumersLen)
	err = sliceCmder.Err()
	if err != nil {
		return nil, fmt.Errorf("(LRANGE pipeKey.consumers).%s", err)
	}
	remainingConsumerNames := sliceCmder.Val()
	remainingConsumers = make(map[string]consumer.Interface)
	for _, consumerName := range remainingConsumerNames {
		if cons, ok := consumers[consumerName]; ok {
			remainingConsumers[consumerName] = cons
		}
	}
	return remainingConsumers, nil
}

func addRedisPipeConsumers(tx *redis.Tx, pipeKey string, consumers map[string]consumer.Interface) (err error) {
	key := fmt.Sprintf("%s.consumers", pipeKey)
	consumerNames := make([]interface{}, 0, len(consumers))
	var consumerName string
	for consumerName = range consumers {
		consumerNames = append(consumerNames, consumerName)
	}
	var intCmder *redis.IntCmd
	intCmder = tx.RPush(key, consumerNames...)
	err = intCmder.Err()
	if err != nil {
		return fmt.Errorf("(RPUSH pipeKey.consumers consumerNames...).%s", err)
	}
	return
}

func deleteRedisPipeConsumer(tx *redis.Tx, pipeKey, consumerName string) (err error) {
	var intCmder *redis.IntCmd
	intCmder = tx.LRem(fmt.Sprintf("%s.consumers", pipeKey), 0, consumerName)
	err = intCmder.Err()
	if err != nil {
		return fmt.Errorf("(LREM pipeKey.consumers consumerName).%s", err)
	}
	return nil
}

func deleteRedisPipeConsumers(tx *redis.Tx, pipeKey string) (err error) {
	var intCmder *redis.IntCmd
	intCmder = tx.Del(fmt.Sprintf("%s.consumers", pipeKey))
	err = intCmder.Err()
	if err != nil {
		return fmt.Errorf("(DEL pipeKey.consumers).%s", err)
	}
	return nil
}
