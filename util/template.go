package util

import (
	configuration "github.com/khezen/espipe/configuration"
	elastic "github.com/khezen/espipe/elastic"
)

// EnsureTemplate creates index if not exist
func EnsureTemplate(client *elastic.Client, template *configuration.Template) error {
	err := client.UpsertTemplate(template)
	if err != nil {
		return err
	}
	return nil
}
