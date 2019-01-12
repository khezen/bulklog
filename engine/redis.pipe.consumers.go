package engine

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/bulklog/bulklog/consumer"
)

func getRedisPipeConsumers(red *redis.Pool, pipeKey string, consumers map[string]consumer.Interface) (remainingConsumers map[string]consumer.Interface, err error) {
	conn := red.Get()
	defer conn.Close()
	key := fmt.Sprintf("%s.consumers", pipeKey)
	remainingConsumersLen, err := conn.Do("LLen", key)
	if err != nil {
		return nil, fmt.Errorf("(LLEN pipeKey.consumers).%s", err)
	}
	if remainingConsumersLen == 0 {
		return map[string]consumer.Interface{}, nil
	}
	remainingConsumerNamesI, err := conn.Do("LRANGE", key, 0, remainingConsumersLen)
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
		consumerName = string(consumerNameI.([]byte))
		if cons, ok := consumers[consumerName]; ok {
			remainingConsumers[consumerName] = cons
		}
	}
	return remainingConsumers, nil
}

func addRedisPipeConsumers(conn redis.Conn, pipeKey string, consumers map[string]consumer.Interface) (err error) {
	key := fmt.Sprintf("%s.consumers", pipeKey)
	args := make([]interface{}, 0, len(consumers)+1)
	args = append(args, key)
	var consumerName string
	for consumerName = range consumers {
		args = append(args, consumerName)
	}
	err = conn.Send("RPUSH", args...)
	if err != nil {
		return fmt.Errorf("(RPUSH pipeKey.consumers consumerNames...).%s", err)
	}
	return
}

func deleteRedisPipeConsumer(red *redis.Pool, pipeKey, consumerName string) (err error) {
	conn := red.Get()
	defer conn.Close()
	_, err = conn.Do("LREM", fmt.Sprintf("%s.consumers", pipeKey), 0, consumerName)
	if err != nil {
		return fmt.Errorf("(LREM pipeKey.consumers consumerName).%s", err)
	}
	return nil
}

func deleteRedisPipeConsumers(conn redis.Conn, pipeKey string) (err error) {
	err = conn.Send("DEL", fmt.Sprintf("%s.consumers", pipeKey))
	if err != nil {
		return fmt.Errorf("(DEL pipeKey.consumers).%s", err)
	}
	return nil
}
