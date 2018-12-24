package elastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/khezen/bulklog/auth"
	"github.com/khezen/bulklog/collection"
	"github.com/khezen/bulklog/consumer"
)

var (
	// ErrNotAcknowledged - creation request has been sent but not acknowledged by elasticsearh
	ErrNotAcknowledged = errors.New("ErrNotAcknowledged - creation request has been sent but not acknowledged by elasticsearh")
)

// Elastic is a client for Elasticsearch API
type Elastic struct {
	signer                         auth.Signer
	indeSettings                   IndexSettings
	bulkEndpoint, templateEndpoint string
}

// New returns a elasticsearch as a consumer
func New(cfg Config) consumer.Interface {
	bulkEndpoint := fmt.Sprintf("%s/_bulk", cfg.Address)
	createTemplateEndpoint := fmt.Sprintf("%s/_template", cfg.Address)
	var signer auth.Signer
	switch {
	case cfg.AWSAuth != nil:
		signer = auth.NewAWSSigner(*cfg.AWSAuth, "es")
		break
	case cfg.BasicAuth != nil:
		signer = auth.NewBasicSigner(*cfg.BasicAuth)
		break
	}
	if cfg.Shards <= 0 {
		cfg.Shards = 1
	}
	return &Elastic{
		signer,
		IndexSettings{
			NumberOfShards: cfg.Shards,
		},
		bulkEndpoint,
		createTemplateEndpoint,
	}
}

// Digest send bulk request to Elasticsearch
func (c *Elastic) Digest(documents []collection.Document) error {
	buf := bytes.NewBuffer([]byte{})
	for _, doc := range documents {
		docBytes, err := Digest(doc)
		if err != nil {
			return err
		}
		buf.Write(docBytes)
	}
	req, err := http.NewRequest("POST", c.bulkEndpoint, buf)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	err = c.sign(req, buf.Bytes())
	if err != nil {
		return err
	}
	_, err = httpClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}

// Ensure creates a template in Elasticsearch
func (c *Elastic) Ensure(collection collection.Collection) error {
	endpoint := fmt.Sprintf("%s/%s", c.templateEndpoint, collection.Name)
	elasticIndex := RenderElasticIndex(collection, c.indeSettings)
	elasticIndexBytes, err := json.Marshal(elasticIndex)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(elasticIndexBytes)
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	err = c.sign(req, elasticIndexBytes)
	if err != nil {
		return err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ErrNotAcknowledged
	}
	return nil
}

func (c *Elastic) sign(req *http.Request, body []byte) (err error) {
	if c.signer != nil {
		err = c.signer.Sign(req, body)
	}
	return err
}
