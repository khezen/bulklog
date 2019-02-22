package engine

import (
	"fmt"

	"github.com/khezen/bulklog/collection"
	"github.com/khezen/bulklog/config"
	"github.com/khezen/bulklog/output"
)

// Indexer indexes document in bulk request to elasticsearch
type engine struct {
	schemas map[collection.Name]map[collection.SchemaName]struct{}
	buffers map[collection.Name]Buffer
}

// New - Create new service for serving web REST requests
func New(cfg *config.Config) (Engine, error) {
	outputs, err := output.NewOutputs(&cfg.Output)
	if err != nil {
		return nil, fmt.Errorf("output.Newoutputs.%s", err)
	}
	schemas := make(map[collection.Name]map[collection.SchemaName]struct{})
	buffers := make(map[collection.Name]Buffer)
	for _, collecCfg := range cfg.Collections {
		collec, err := collection.New(collecCfg)
		if err != nil {
			return nil, fmt.Errorf("collection.New.%s", err)
		}
		for _, cons := range outputs {
			err = cons.Ensure(collec)
			if err != nil {
				return nil, fmt.Errorf("Ensure.%s", err)
			}
		}
		schemas[collec.Name] = make(map[collection.SchemaName]struct{})
		for _, schema := range collec.Schemas {
			schemas[collec.Name][schema.Name] = struct{}{}
		}
		var buffer Buffer
		if cfg.Persistence.Enabled {
			buffer = RedisBuffer(collec, &cfg.Persistence.Redis, outputs)
		} else {
			buffer = DefaultBuffer(collec, outputs)
		}
		buffers[collec.Name] = buffer
		if collec.FlushPeriod > 0 {
			go buffer.Flusher()()
		}
	}
	return &engine{
		schemas,
		buffers,
	}, nil
}

// Collect document
func (e *engine) Collect(collectionName collection.Name, schemaName collection.SchemaName, docBytes []byte) (err error) {
	_, ok := e.schemas[collectionName]
	if !ok {
		return ErrNotFound
	}
	if _, ok := e.schemas[collectionName][schemaName]; !ok {
		return ErrNotFound
	}
	document, err := collection.NewDocument(collectionName, schemaName, docBytes)
	if err != nil {
		return fmt.Errorf("collection.NewDocument.%s", err)
	}
	err = e.Dispatch(document)
	if err != nil {
		return fmt.Errorf("Dispatch.%s", err)
	}
	return nil
}

// Dispatch takes incoming message into Elasticsearch
func (e *engine) Dispatch(document *collection.Document) (err error) {
	err = e.buffers[document.CollectionName].Append(document)
	if err != nil {
		return fmt.Errorf("Append.%s", err)
	}
	return nil
}

// Collect document
func (e *engine) CollectBatch(collectionName collection.Name, schemaName collection.SchemaName, docBytesSlice ...[]byte) (err error) {
	_, ok := e.schemas[collectionName]
	if !ok {
		return ErrNotFound
	}
	if _, ok := e.schemas[collectionName][schemaName]; !ok {
		return ErrNotFound
	}
	length := len(docBytesSlice)
	if length > 0 {
		documents := make([]collection.Document, 0, length)
		var docBytes []byte
		for _, docBytes = range docBytesSlice {
			document, err := collection.NewDocument(collectionName, schemaName, docBytes)
			if err != nil {
				return fmt.Errorf("collection.NewDocument.%s", err)
			}
			documents = append(documents, *document)
		}
		err = e.DispatchBatch(documents...)
		if err != nil {
			return fmt.Errorf("Dispatch.%s", err)
		}
	}
	return nil
}

// Dispatch takes incoming message into Elasticsearch
func (e *engine) DispatchBatch(documents ...collection.Document) (err error) {
	if len(documents) > 0 {
		err = e.buffers[documents[0].CollectionName].AppendBatch(documents...)
		if err != nil {
			return fmt.Errorf("Append.%s", err)
		}
	}
	return nil
}
