package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/khezen/espipe/collection"
	"github.com/khezen/espipe/consumer"
)

const bufferLimit = 10000

// buffer is related to a template
// It sends messages in bulk to elasticsearch.
type buffer struct {
	sync.Mutex
	collection collection.Collection
	consumers  []consumer.Interface
	close      chan struct{}
	documents  []collection.Document
}

// DefaultBuffer creates a new buffer
func DefaultBuffer(collec collection.Collection, consumers ...consumer.Interface) Buffer {
	buffer := &buffer{
		Mutex:      sync.Mutex{},
		collection: collec,
		consumers:  consumers,
		close:      make(chan struct{}),
		documents:  make([]collection.Document, 0),
	}
	return buffer
}

func (b *buffer) Set(consumers ...consumer.Interface) {
	b.Lock()
	b.consumers = consumers
	b.Unlock()
}

// Append to buffer
func (b *buffer) Append(d collection.Document) error {
	b.Lock()
	b.documents = append(b.documents, d)
	b.Unlock()
	return nil
}

// Flush the buffer
func (b *buffer) Flush() (bubbledErr error) {
	b.Lock()
	defer b.Unlock()
	documentsLen := len(b.documents)
	if documentsLen == 0 {
		return nil
	}
	documentsCpy := b.documents
	b.documents = make([]collection.Document, 0, bufferLimit)
	consumersCpy := append(make([]consumer.Interface, 0, len(b.consumers)), b.consumers...)
	go convey(documentsCpy, consumersCpy, b.collection.FlushPeriod, b.collection.RetentionPeriod)
	return nil
}

// Flusher flushes every tick
func (b *buffer) Flusher() func() {
	return func() {
		var (
			ticker = time.NewTicker(b.collection.FlushPeriod)
			err    error
		)
		for {
			select {
			case <-b.close:
				return
			case <-ticker.C:
				err = b.Flush()
				if err != nil {
					fmt.Println(err)
				}
				break
			}
		}
	}
}

func (b *buffer) Close() {
	b.close <- struct{}{}
}
