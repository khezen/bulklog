package configuration

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {

	cases := []struct {
		filePath  string
		expectErr bool
	}{
		{"testValidConfig.json", false},
		{"notFound.json", true},
		{"testUnparsable.json", true},
	}

	for _, c := range cases {
		_, err := LoadConfig(c.filePath)
		switch {
		case c.expectErr && err == nil:
			t.Error("Expected error got nil")
		case !c.expectErr && err != nil:
			panic(err)
		}
	}
}

func TestGetTypes(t *testing.T) {
	config, err := LoadConfig("testValidConfig.json")
	if err != nil {
		panic(err)
	}
	wrongMappingsConf, err := LoadConfig("testMappingCastFail.json")
	if err != nil {
		panic(err)
	}
	cases := []struct {
		template      *Template
		expectedTypes []string
		expectErr     bool
	}{
		{&config.Templates[0], []string{"log"}, false},
		{&config.Templates[1], []string{"trace"}, false},
		{&wrongMappingsConf.Templates[0], []string{"log"}, true},
	}
	for _, c := range cases {
		types, err := c.template.GetTypes()
		switch {
		case c.expectErr && err == nil:
			t.Error("Expected error, got nil")
			break
		case !c.expectErr && err != nil:
			panic(err)
		}
		if c.expectErr {
			continue
		}
		if len(types) != len(c.expectedTypes) {
			t.Errorf("len(GetTypes(index)): Expected %v, Got %v", len(c.expectedTypes), len(types))
		}
		for _, toBeFound := range types {
			found := false
			for _, current := range c.expectedTypes {
				if current == toBeFound {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected %v, Got %v", c.expectedTypes, types)
			}
		}
	}
}
