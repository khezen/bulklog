package collection

import (
	"fmt"
	"time"
)

// New Collection
func New(cfg Config) (*Collection, error) {
	flushPeriod, err := cfg.FlushPeriod()
	if err != nil {
		return nil, fmt.Errorf("FlushPeriod.%s", err)
	}
	retentionPeriod, err := cfg.RetentionPeriod()
	if err != nil {
		return nil, fmt.Errorf("RetnetionPeriod.%s", err)
	}
	schemas, err := cfg.Schemas()
	if err != nil {
		return nil, fmt.Errorf("Schemas.%s", err)
	}
	return &Collection{
		Name:            cfg.Name,
		FlushPeriod:     flushPeriod,
		RetentionPeriod: retentionPeriod,
		Schemas:         schemas,
	}, nil
}

// Collection descrbies a document Template
type Collection struct {
	Name            Name
	FlushPeriod     time.Duration
	RetentionPeriod time.Duration
	Schemas         []Schema
}

// Name of a collection
type Name string

// Schema - document schema
type Schema struct {
	Name   SchemaName
	Fields map[string]Field
}

// SchemaName -
type SchemaName string

// Field -
type Field struct {
	Type       FieldType `yaml:"type"`
	Length     int       `yaml:"length"`
	MaxLength  int       `yaml:"max_length"`
	DateFormat string    `yaml:"date_format"`
}

// FieldType -
type FieldType string
