package collection

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Document has a JSON body which must be indexed in given Template as given Type.
type Document struct {
	CollectionName Name
	SchemaName     SchemaName
	ID             uuid.UUID
	PostedAt       time.Time
	Body           []byte
}

// NewDocument creates a document from es index, document type and its body
func NewDocument(collectionName Name, schemaName SchemaName, body []byte) (Document, error) {
	now := time.Now().UTC()
	id := uuid.New()
	var bodyMap map[string]interface{}
	err := json.Unmarshal(body, &bodyMap)
	if err != nil {
		return Document{}, err
	}
	body, err = json.Marshal(bodyMap)
	if err != nil {
		return Document{}, err
	}
	return Document{
		ID:             id,
		PostedAt:       now,
		CollectionName: collectionName,
		SchemaName:     schemaName,
		Body:           body,
	}, nil
}
