package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/khezen/bulklog/collection"
)

func setRedisDocuments(tx redis.Pipeliner, bufferKey, pipeKey string) (err error) {
	_, err = tx.Rename(bufferKey, fmt.Sprintf("%s.buffer", pipeKey)).Result()
	return err
}

func getRedisDocuments(red *redis.Client, pipeKey string) (documents []collection.Document, err error) {
	bufferKey := fmt.Sprintf("%s.buffer", pipeKey)
	documentsLen, err := red.LLen(bufferKey).Result()
	if err != nil {
		return nil, err
	}
	docStrings, err := red.LRange(bufferKey, 0, documentsLen).Result()
	if err != nil {
		return nil, err
	}
	documents = make([]collection.Document, 0, documentsLen)
	var buf *bytes.Buffer
	for _, docBase64 := range docStrings {
		docBytes, err := base64.StdEncoding.DecodeString(docBase64)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(docBytes)
		var doc collection.Document
		err = gob.NewDecoder(buf).Decode(&doc)
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func deleteRedisDocuments(tx redis.Pipeliner, pipeKey string) (err error) {
	_, err = tx.Del(fmt.Sprintf("%s.buffer", pipeKey)).Result()
	return err
}
