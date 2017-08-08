package dispatcher

import (
	configuration "github.com/khezen/espipe/configuration"
	elastic "github.com/khezen/espipe/elastic"
	model "github.com/khezen/espipe/model"
	"sync"
)

// Dispatcher dispatch logs message to Elasticsearch
type Dispatcher struct {
	Client  *elastic.Client
	buffers map[configuration.TemplateName]*Buffer
	relieve chan *Buffer
	mutex   sync.RWMutex
}

// NewDispatcher creates a new Dispatcher object
func NewDispatcher(config *configuration.Configuration) (*Dispatcher, error) {
	buffers := make(map[configuration.TemplateName]*Buffer)
	client := elastic.NewClient(config)
	return &Dispatcher{
		client,
		buffers,
		make(chan *Buffer),
		sync.RWMutex{},
	}, nil
}

// Dispatch takes incoming message into Elasticsearch
func (d *Dispatcher) Dispatch(document *model.Document) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.ensureBuffer(document)
	d.buffers[document.Template.Name].Append <- *document
}

func (d *Dispatcher) ensureBuffer(document *model.Document) {
	if _, ok := d.buffers[document.Template.Name]; !ok {
		buffer := NewBuffer(document.Template, d.Client)
		go buffer.Gophers()()
		d.buffers[document.Template.Name] = buffer
	}
}
