package engine

import (
	"github.com/bulklog/bulklog/collection"
)

// Engine -
type Engine interface {
	Collector
	Dispatcher
}

// Dispatcher dispatches documents
type Dispatcher interface {
	Dispatch(document *collection.Document) error
	DispatchBatch(documents ...collection.Document) error
}

// Collector collects documents
type Collector interface {
	Collect(collectionName collection.Name, schemaName collection.SchemaName, docBytes []byte) error
	CollectBatch(collectionName collection.Name, schemaName collection.SchemaName, docBytesSlice ...[]byte) error
}

// Buffer -
type Buffer interface {
	Append(*collection.Document) error
	AppendBatch(...collection.Document) error
	Flush() error
	Flusher() func()

	Close()
}
