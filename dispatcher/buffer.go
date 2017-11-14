package dispatcher

import (
	"fmt"
	"sync"
	"time"

	configuration "github.com/khezen/espipe/configuration"
	elastic "github.com/khezen/espipe/elastic"
	model "github.com/khezen/espipe/model"
)

// Buffer is related to a template
// It sends messages in bulk to elasticsearch.
type Buffer struct {
	Template  *configuration.Template
	client    *elastic.Client
	Append    chan model.Document
	Kill      chan error
	documents []model.Document
	sizeKB    float64
	mutex     sync.RWMutex
}

const bufferLimit = 1000

// NewBuffer creates a new buffer
func NewBuffer(template *configuration.Template, client *elastic.Client) *Buffer {
	return &Buffer{
		template,
		client,
		make(chan model.Document),
		make(chan error),
		make([]model.Document, 0),
		0,
		sync.RWMutex{},
	}
}

func (b *Buffer) append(msg model.Document) {
	b.mutex.Lock()
	b.documents = append(b.documents, msg)
	b.sizeKB += float64(len(msg.Body)) / 1000
	if b.sizeKB >= b.Template.BufferSizeKB || len(b.documents) >= bufferLimit {
		b.mutex.Unlock()
		go b.flush()
	} else {
		b.mutex.Unlock()
	}
}

func (b *Buffer) flush() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if len(b.documents) == 0 {
		return
	}
	bulk := make([]byte, 0, int(b.sizeKB)+len(b.documents)*150)
	for _, doc := range b.documents {
		req, err := doc.Request()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		bulk = append(bulk, req...)
	}
	err := b.client.Bulk(bulk)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	b.documents = make([]model.Document, 0, bufferLimit)
	b.sizeKB = 0
}

// Gophers returns a func() in wich go routines are taking new message to the bulk.
func (b *Buffer) Gophers() func() {
	ticker := time.NewTicker(time.Duration(b.Template.TimerMS) * time.Millisecond)
	return func() {
		for {
			select {
			case <-b.Kill:
				return
			case <-ticker.C:
				go b.flush()
				break
			case msg := <-b.Append:
				b.append(msg)
				break
			}
		}
	}
}
