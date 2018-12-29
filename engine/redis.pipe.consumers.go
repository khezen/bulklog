package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	"github.com/go-redis/redis"
	"github.com/khezen/bulklog/consumer"
)

func getRedisConsumers(tx redis.Pipeliner, pipeKey string) (consumers []consumer.Interface, err error) {
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

func setRedisConsumers(tx redis.Pipeliner, pipeKey string, consumers []consumer.Interface) (err error) {
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(consumers)
	if err != nil {
		return err
	}
	consumersBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	_, err = tx.HSet(pipeKey, "consumers", consumersBase64).Result()
	if err != nil {
		return err
	}
	return nil
}
