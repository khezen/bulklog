package model

import (
	"github.com/khezen/espipe/configuration"
	"io/ioutil"
	"testing"
)

func TestNewDocument(t *testing.T) {
	config, err := configuration.LoadConfig("../configuration/testValidConfig.json")
	logTemplate := &config.Templates[0]
	if err != nil {
		panic(err)
	}

	fullLog, err := ioutil.ReadFile("testResources/fullLog.json")
	if err != nil {
		panic(err)
	}
	missingServiceLog, err := ioutil.ReadFile("testResources/missingService.json")
	if err != nil {
		panic(err)
	}

	missngTimestampLog, err := ioutil.ReadFile("testResources/missingTimestamp.json")
	if err != nil {
		panic(err)
	}

	missingBoth, err := ioutil.ReadFile("testResources/missingBoth.json")
	if err != nil {
		panic(err)
	}

	unparsable, err := ioutil.ReadFile("testResources/unparsable.json")
	if err != nil {
		panic(err)
	}

	cases := []struct {
		index     *configuration.Template
		docType   DocumentType
		body      []byte
		expectErr bool
	}{
		{logTemplate, "log", fullLog, false},
		{logTemplate, "log", missingServiceLog, false},
		{logTemplate, "log", missngTimestampLog, false},
		{logTemplate, "log", missingBoth, false},
		{logTemplate, "log", unparsable, true},
	}
	for _, c := range cases {
		_, err := NewDocument(c.index, c.docType, c.body)
		switch {
		case c.expectErr && err == nil:
			t.Error("Expected error got nil")
		case !c.expectErr && err != nil:
			panic(err)
		}
	}
}
