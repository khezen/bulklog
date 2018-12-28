package collection

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Document has a JSON body which must be indexed in given Template as given Type.
type Document struct {
	ID             uuid.UUID
	PostedAt       time.Time
	CollectionName Name
	SchemaName     SchemaName
	Body           []byte
}

// NewDocument creates a document from es index, document type and its body
func NewDocument(collectionName Name, schemaName SchemaName, body []byte) (*Document, error) {
	var bodyMap map[string]interface{}
	err := json.Unmarshal(body, &bodyMap)
	if err != nil {
		return nil, err
	}
	body, err = json.Marshal(bodyMap)
	if err != nil {
		return nil, err
	}
	return &Document{
		ID:             uuid.New(),
		PostedAt:       time.Now().UTC(),
		CollectionName: collectionName,
		SchemaName:     schemaName,
		Body:           body,
	}, nil
}
