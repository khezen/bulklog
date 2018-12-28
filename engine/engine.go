package engine

import (
	"github.com/khezen/bulklog/collection"
	"github.com/khezen/bulklog/config"
	"github.com/khezen/bulklog/consumer"
	"github.com/khezen/bulklog/consumer/elastic"
)

// Indexer indexes document in bulk request to elasticsearch
type engine struct {
	schemas map[collection.Name]map[collection.SchemaName]struct{}
	buffers map[collection.Name]Buffer
}

// New - Create new service for serving web REST requests
func New(cfg *config.Config) (Engine, error) {
	consumers := make([]consumer.Interface, 0, 5)
	if cfg.Output.Elastic != nil {
		elasticsearch := elastic.New(*cfg.Output.Elastic)
		consumers = append(consumers, elasticsearch)
	}
	schemas := make(map[collection.Name]map[collection.SchemaName]struct{})
	buffers := make(map[collection.Name]Buffer)
	for _, collecCfg := range cfg.Collections {
		collec, err := collection.New(collecCfg)
		if err != nil {
			return nil, err
		}
		for _, cons := range consumers {
			err = cons.Ensure(collec)
			if err != nil {
				return nil, err
			}
		}
		schemas[collec.Name] = make(map[collection.SchemaName]struct{})
		for _, schema := range collec.Schemas {
			schemas[collec.Name][schema.Name] = struct{}{}
		}
		var buffer Buffer
		if cfg.Persistence.Enabled {
			buffer, err = RedisBuffer(collec, cfg.Persistence.Redis, consumers...)
			if err != nil {
				return nil, err
			}
		} else {
			buffer = DefaultBuffer(collec, consumers...)
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
func (e *engine) Collect(collectionName collection.Name, schemaName collection.SchemaName, docBytes []byte) error {
	_, ok := e.schemas[collectionName]
	if !ok {
		return ErrNotFound
	}
	if _, ok := e.schemas[collectionName][schemaName]; !ok {
		return ErrNotFound
	}
	document, err := collection.NewDocument(collectionName, schemaName, docBytes)
	if err != nil {
		return err
	}
	return e.Dispatch(document)
}

// Dispatch takes incoming message into Elasticsearch
func (e *engine) Dispatch(document *collection.Document) error {
	return e.buffers[document.CollectionName].Append(document)
}
