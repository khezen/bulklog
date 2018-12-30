package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/khezen/bulklog/collection"
)

func flushBuffer2RedisPipe(tx *redis.Tx, bufferKey, pipeKey string) (err error) {
	var statusCmder *redis.StatusCmd
	statusCmder = tx.Rename(bufferKey, fmt.Sprintf("%s.buffer", pipeKey))
	err = statusCmder.Err()
	if err != nil {
		return fmt.Errorf("RENAME(bufferKey pipeKey.buffer).%s", err)
	}
	return nil
}

func getRedisPipeDocuments(red *redis.Client, pipeKey string) (documents []collection.Document, err error) {
	var (
		intCmder   *redis.IntCmd
		sliceCmder *redis.StringSliceCmd
	)
	bufferKey := fmt.Sprintf("%s.buffer", pipeKey)
	intCmder = red.LLen(bufferKey)
	err = intCmder.Err()
	if err != nil {
		return nil, fmt.Errorf("(LLEN pipeKey.buffer).%s", err)
	}
	documentsLen := intCmder.Val()
	sliceCmder = red.LRange(bufferKey, 0, documentsLen)
	err = sliceCmder.Err()
	if err != nil {
		return nil, fmt.Errorf("(LRANGE pipeKey.buffer 0 documentsLen).%s", err)
	}
	docStrings := sliceCmder.Val()
	documents = make([]collection.Document, 0, documentsLen)
	var buf *bytes.Buffer
	for _, docBase64 := range docStrings {
		docBytes, err := base64.StdEncoding.DecodeString(docBase64)
		if err != nil {
			return nil, fmt.Errorf("base64.std.decode.%s", err)
		}
		buf = bytes.NewBuffer(docBytes)
		var doc collection.Document
		err = gob.NewDecoder(buf).Decode(&doc)
		if err != nil {
			return nil, fmt.Errorf("(gob.decode.%s", err)
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func deleteRedisPipeDocuments(tx *redis.Tx, pipeKey string) (err error) {
	var intCmder *redis.IntCmd
	intCmder = tx.Del(fmt.Sprintf("%s.buffer", pipeKey))
	err = intCmder.Err()
	if err != nil {
		return fmt.Errorf("(DEL pipeKey.buffer).%s", err)
	}
	return nil
}
