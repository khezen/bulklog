package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/khezen/espipe/collection"
)

func getDocuments(red redis.Client, pipeKey string) (documents []collection.Document, err error) {
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
