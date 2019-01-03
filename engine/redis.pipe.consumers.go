package engine

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/khezen/bulklog/consumer"
)

func getRedisPipeConsumers(red redis.Conn, pipeKey string, consumers map[string]consumer.Interface) (remainingConsumers map[string]consumer.Interface, err error) {
	key := fmt.Sprintf("%s.consumers", pipeKey)
	remainingConsumersLen, err := red.Do("LLen", key)
	if err != nil {
		return nil, fmt.Errorf("(LLEN pipeKey.consumers).%s", err)
	}
	if remainingConsumersLen == 0 {
		return map[string]consumer.Interface{}, nil
	}
	remainingConsumerNamesI, err := red.Do("Range", key, 0, remainingConsumersLen)
	if err != nil {
		return nil, fmt.Errorf("(LRANGE pipeKey.consumers).%s", err)
	}
	remainingConsumerNames := remainingConsumerNamesI.([]interface{})
	remainingConsumers = make(map[string]consumer.Interface)
	var (
		consumerNameI interface{}
		consumerName  string
	)
	for _, consumerNameI = range remainingConsumerNames {
		consumerName = consumerNameI.(string)
		if cons, ok := consumers[consumerName]; ok {
			remainingConsumers[consumerName] = cons
		}
	}
	return remainingConsumers, nil
}

func addRedisPipeConsumers(red redis.Conn, pipeKey string, consumers map[string]consumer.Interface) (err error) {
	key := fmt.Sprintf("%s.consumers", pipeKey)
	args := make([]interface{}, 0, len(consumers)+1)
	args = append(args, key)
	var consumerName string
	for consumerName = range consumers {
		args = append(args, consumerName)
	}
	err = red.Send("RPUSH", args...)
	if err != nil {
		return fmt.Errorf("(RPUSH pipeKey.consumers consumerNames...).%s", err)
	}
	return
}

func deleteRedisPipeConsumer(red redis.Conn, pipeKey, consumerName string) (err error) {
	err = red.Send("LREM", fmt.Sprintf("%s.consumers", pipeKey), 0, consumerName)
	if err != nil {
		return fmt.Errorf("(LREM pipeKey.consumers consumerName).%s", err)
	}
	return nil
}

func deleteRedisPipeConsumers(red redis.Conn, pipeKey string) (err error) {
	err = red.Send("DEL", fmt.Sprintf("%s.consumers", pipeKey))
	if err != nil {
		return fmt.Errorf("(DEL pipeKey.consumers).%s", err)
	}
	return nil
}
