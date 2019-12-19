package output

import "github.com/khezen/bulklog/pkg/collection"

// Interface interface to send msg to recipents
type Interface interface {
	Digest(documents []collection.Document) error
	Ensure(collection *collection.Collection) error
}
