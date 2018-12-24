package engine

import (
	"github.com/khezen/bulklog/collection"
	"github.com/khezen/bulklog/consumer"
)

// Engine -
type Engine interface {
	Collector
	Dispatcher
}

// Dispatcher dispatches documents
type Dispatcher interface {
	Dispatch(document collection.Document) error
}

// Collector collects documents
type Collector interface {
	Collect(collectionName collection.Name, schemaName collection.SchemaName, docBytes []byte) error
}

// Buffer -
type Buffer interface {
	Set(...consumer.Interface)
	Append(collection.Document) error
	Flush() error
	Flusher() func()

	Close()
}
