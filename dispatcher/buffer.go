package dispatcher

import (
	"fmt"
	configuration "github.com/khezen/espipe/configuration"
	elastic "github.com/khezen/espipe/elastic"
	model "github.com/khezen/espipe/model"
	"sync"
	"time"
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

// NewBuffer creates a new buffer
func NewBuffer(template *configuration.Template, client *elastic.Client) *Buffer {
	return &Buffer{
		template,
		client,
		make(chan model.Document),
		make(chan error),
		make([]model.Document, 0, template.BufferLen),
		0,
		sync.RWMutex{},
	}
}

func (b *Buffer) append(msg model.Document) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.documents = append(b.documents, msg)
	b.sizeKB += float64(len(msg.Body)) / 1000
	if b.sizeKB >= b.Template.BufferSizeKB || len(b.documents) >= b.Template.BufferLen {
		err := b.flush()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (b *Buffer) flush() error {
	if len(b.documents) == 0 {
		return nil
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	bulk := make([]byte, 0, int(b.sizeKB)+len(b.documents)*150)
	for _, doc := range b.documents {
		req, err := doc.Request()
		if err != nil {
			return err
		}
		bulk = append(bulk, req...)
	}
	err := b.client.Bulk(bulk)
	if err != nil {
		return err
	}
	b.documents = make([]model.Document, 0, b.Template.BufferLen)
	b.sizeKB = 0
	return nil
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
				err := b.flush()
				if err != nil {
					fmt.Println(err)
				}
				break
			case msg := <-b.Append:
				b.append(msg)
				break
			}
		}
	}
}
