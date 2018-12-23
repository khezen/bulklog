package auth

import (
	"bytes"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	signer "github.com/aws/aws-sdk-go/aws/signer/v4"
)

// AWSConfig provide credential for AWS services signing
type AWSConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
}

// NewAWSSigner provides  AWS v4 signing to given request
func NewAWSSigner(cfg AWSConfig, service string) Signer {
	cred := credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	return &awsSigner{
		service,
		cfg.Region,
		signer.NewSigner(cred),
	}
}

type awsSigner struct {
	service string
	region  string
	client  *signer.Signer
}

func (s *awsSigner) Sign(req *http.Request, body []byte) error {
	byteReader := bytes.NewReader(body)
	_, err := s.client.Sign(req, byteReader, "es", s.region, time.Now())
	return err
}
