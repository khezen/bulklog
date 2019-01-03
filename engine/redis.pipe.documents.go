package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/khezen/bulklog/collection"
)

func flushBuffer2RedisPipe(red redis.Conn, bufferKey, pipeKey string) (err error) {
	err = red.Send("RENAME", bufferKey, fmt.Sprintf("%s.buffer", pipeKey))
	if err != nil {
		return fmt.Errorf("RENAME(bufferKey pipeKey.buffer).%s", err)
	}
	return nil
}

func getRedisPipeDocuments(red redis.Conn, pipeKey string) (documents []collection.Document, err error) {
	bufferKey := fmt.Sprintf("%s.buffer", pipeKey)
	documentsLenI, err := red.Do("LLEN", bufferKey)
	if err != nil {
		return nil, fmt.Errorf("(LLEN pipeKey.buffer).%s", err)
	}
	documentsLen := documentsLenI.(int64)
	if documentsLen == 0 {
		return []collection.Document{}, nil
	}
	docStringsI, err := red.Do("LRANGE", bufferKey, 0, documentsLen)
	if err != nil {
		return nil, fmt.Errorf("(LRANGE pipeKey.buffer 0 documentsLen).%s", err)
	}
	docStrings := docStringsI.([]interface{})
	documents = make([]collection.Document, 0, documentsLen)
	var buf *bytes.Buffer
	for _, docBase64 := range docStrings {
		docBytes, err := base64.StdEncoding.DecodeString(docBase64.(string))
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

func deleteRedisPipeDocuments(red redis.Conn, pipeKey string) (err error) {
	err = red.Send("DEL", fmt.Sprintf("%s.buffer", pipeKey))
	if err != nil {
		return fmt.Errorf("(DEL pipeKey.buffer).%s", err)
	}
	return nil
}
