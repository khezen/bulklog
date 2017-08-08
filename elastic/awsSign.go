package elastic

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	signer "github.com/aws/aws-sdk-go/aws/signer/v4"
)

// NewAWSSigner provides  AWS v4 signing to given request
func NewAWSSigner(accessKeyID, secretAccessKey string) *signer.Signer {
	cred := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
	return signer.NewSigner(cred)
}
