package elastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bulklog/bulklog/auth"
	"github.com/bulklog/bulklog/collection"
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
	httpcli                        http.Client
}

// New returns a elasticsearch as a output
func New(cfg Config) *Elastic {
	if cfg.Scheme == "" {
		cfg.Scheme = "http"
	}
	bulkEndpoint := fmt.Sprintf("%s://%s/_bulk", cfg.Scheme, cfg.Endpoint)
	createTemplateEndpoint := fmt.Sprintf("%s://%s/_template", cfg.Scheme, cfg.Endpoint)
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
		http.Client{
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: true,
			},
		},
	}
}

// Digest send bulk request to Elasticsearch
func (c *Elastic) Digest(documents []collection.Document) error {
	buf := bytes.NewBuffer([]byte{})
	for _, doc := range documents {
		docBytes, err := Digest(doc)
		if err != nil {
			return fmt.Errorf("Digest.%s", err)
		}
		buf.Write(docBytes)
	}
	req, err := http.NewRequest("POST", c.bulkEndpoint, buf)
	if err != nil {
		return fmt.Errorf("http.NewRequest.%s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	err = c.sign(req, buf.Bytes())
	if err != nil {
		return fmt.Errorf("Sign.%s", err)
	}
	res, err := c.httpcli.Do(req)
	if err != nil {
		return fmt.Errorf("httpClient.Do.%s", err)
	}
	if res.StatusCode > 300 {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("ioutil.ReadAll.%s", err)
		}
		return fmt.Errorf("elasticsearch: %s : %b", res.Status, resBody)
	}
	return nil
}

// Ensure creates a template in Elasticsearch
func (c *Elastic) Ensure(collection *collection.Collection) error {
	endpoint := fmt.Sprintf("%s/%s", c.templateEndpoint, collection.Name)
	elasticIndex := RenderElasticIndex(collection, c.indeSettings)
	elasticIndexBytes, err := json.Marshal(elasticIndex)
	if err != nil {
		return fmt.Errorf("json.Marshal.%s", err)
	}
	buf := bytes.NewBuffer(elasticIndexBytes)
	req, err := http.NewRequest("POST", endpoint, buf)
	if err != nil {
		return fmt.Errorf("http.NewRequest.%s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	err = c.sign(req, elasticIndexBytes)
	if err != nil {
		return fmt.Errorf("Sign.%s", err)
	}
	res, err := c.httpcli.Do(req)
	if err != nil {
		return fmt.Errorf("httpClient.Do.%s", err)
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
