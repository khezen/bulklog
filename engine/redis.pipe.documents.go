package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/khezen/bulklog/collection"
)

func flushBuffer2RedisPipe(tx redis.Pipeliner, bufferKey, pipeKey string) (err error) {
	_, err = tx.Rename(bufferKey, fmt.Sprintf("%s.buffer", pipeKey)).Result()
	if err != nil {
		return fmt.Errorf("RENAME(bufferKey pipeKey.buffer).%s", err.Error())
	}
	return nil
}

func getRedisPipeDocuments(red *redis.Client, pipeKey string) (documents []collection.Document, err error) {
	bufferKey := fmt.Sprintf("%s.buffer", pipeKey)
	documentsLen, err := red.LLen(bufferKey).Result()
	if err != nil {
		return nil, fmt.Errorf("(LLEN pipeKey.buffer).%s", err.Error())
	}
	docStrings, err := red.LRange(bufferKey, 0, documentsLen).Result()
	if err != nil {
		return nil, fmt.Errorf("(LRANGE pipeKey.buffer 0 documentsLen).%s", err.Error())
	}
	documents = make([]collection.Document, 0, documentsLen)
	var buf *bytes.Buffer
	for _, docBase64 := range docStrings {
		docBytes, err := base64.StdEncoding.DecodeString(docBase64)
		if err != nil {
			return nil, fmt.Errorf("base64.std.decode.%s", err.Error())
		}
		buf = bytes.NewBuffer(docBytes)
		var doc collection.Document
		err = gob.NewDecoder(buf).Decode(&doc)
		if err != nil {
			return nil, fmt.Errorf("(gob.decode.%s", err.Error())
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func deleteRedisPipeDocuments(tx redis.Pipeliner, pipeKey string) (err error) {
	_, err = tx.Del(fmt.Sprintf("%s.buffer", pipeKey)).Result()
	if err != nil {
		return fmt.Errorf("(DEL pipeKey.buffer).%s", err.Error())
	}
	return nil
}
