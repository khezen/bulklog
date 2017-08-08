package elastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	AWSSigner "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/khezen/espipe/configuration"
	"github.com/khezen/espipe/httpcli"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	httpClient = httpcli.Singleton()
	// ErrNotAcknowledged - creation request has been sent but not acknowledged by elasticsearh
	ErrNotAcknowledged = errors.New("ErrNotAcknowledged - creation request has been sent but not acknowledged by elasticsearh")
)

// Client is a client for Elasticsearch API
type Client struct {
	config                         *configuration.Configuration
	AWSSigner                      *AWSSigner.Signer
	BasicAuthSigner                *BasicAuthSigner
	bulkEndpoint, templateEndpoint string
}

// NewClient returns a client for Elasticsearch API
func NewClient(config *configuration.Configuration) *Client {

	bufBulkEndP := bytes.NewBufferString(config.Elasticsearch)
	bufBulkEndP.WriteString("/_bulk")
	bufCreateTemplateEndP := bytes.NewBufferString(config.Elasticsearch)
	bufCreateTemplateEndP.WriteString("/_template/")

	var AWSSigner *AWSSigner.Signer
	var basicAuthSigner *BasicAuthSigner
	switch {
	case config.AWSAuth != nil:
		AWSSigner = NewAWSSigner(config.AWSAuth.AccessKeyID, config.AWSAuth.SecretAccessKey)
		break
	case config.BasicAuth != nil:
		basicAuthSigner = NewBasicAuthSigner(config.BasicAuth.Username, config.BasicAuth.Password)
		break
	}

	return &Client{
		config,
		AWSSigner,
		basicAuthSigner,
		bufBulkEndP.String(),
		bufCreateTemplateEndP.String(),
	}
}

// Bulk send bulk request to Elasticsearch
func (c *Client) Bulk(requestBody []byte) error {
	bodyReader := bytes.NewReader(requestBody)
	req, err := http.NewRequest("POST", c.bulkEndpoint, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	err = c.sign(req, bodyReader)
	if err != nil {
		return err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Printf("\n%v\n%v\n%v%v\n%v\n", time.Now(), c.bulkEndpoint, bytes.NewBuffer(requestBody).String(), res.Status, bytes.NewBuffer(resBody).String())
	return nil
}

// UpsertTemplate creates a template in Elasticsearch
func (c *Client) UpsertTemplate(t *configuration.Template) error {
	endpoint := c.renderTemplateURI(t)
	requestBody, err := json.Marshal(t.Body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(requestBody)
	req, err := http.NewRequest("POST", endpoint, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	err = c.sign(req, bodyReader)
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

func (c *Client) renderTemplateURI(t *configuration.Template) string {
	bufEndpoint := bytes.NewBufferString(c.templateEndpoint)
	bufEndpoint.WriteString(string(t.Name))
	return bufEndpoint.String()
}

func (c *Client) sign(req *http.Request, bodyReader io.ReadSeeker) error {
	var err error
	switch {
	case c.AWSSigner != nil:
		_, err = c.AWSSigner.Sign(req, bodyReader, "es", c.config.AWSAuth.Region, time.Now())
		break
	case c.BasicAuthSigner != nil:
		_, err = c.BasicAuthSigner.Sign(req)
		break
	}
	return err
}
