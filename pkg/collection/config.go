package collection

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Config -
type Config struct {
	Name               Name         `yaml:"name"`
	FlushPeriodStr     string       `yaml:"flush_period"`
	RetentionPeriodStr string       `yaml:"retention_period"`
	NumberOfShards     int          `yaml:"shards"`
	NumberOfReplicas   int          `yaml:"replicas"`
	SchemaCfg          SchemaConfig `yaml:"schema"`
}

// SchemaConfig -
type SchemaConfig map[string]Field

// FlushPeriod - extract flush period from config
func (c *Config) FlushPeriod() (time.Duration, error) {
	if c.FlushPeriodStr == "" {
		return 0, nil
	}
	return period(c.FlushPeriodStr)
}

// RetentionPeriod - extract flush period from config
func (c *Config) RetentionPeriod() (time.Duration, error) {
	if c.RetentionPeriodStr == "" {
		return 0, nil
	}
	return period(c.RetentionPeriodStr)
}

func period(periodStr string) (period time.Duration, err error) {
	periodStrSplit := strings.Split(periodStr, " ")
	if len(periodStrSplit) != 2 {
		return period, ErrWrongPeriod
	}
	quantity, err := strconv.ParseFloat(periodStrSplit[0], 64)
	if err != nil {
		return period, fmt.Errorf("strconv.ParseFloat.%s", err)
	}
	unit := strings.ToLower(strings.TrimSpace(periodStrSplit[1]))
	switch unit {
	case "hours":
		period = time.Duration(quantity * float64(time.Hour))
	case "minutes":
		period = time.Duration(quantity * float64(time.Minute))
	case "seconds":
		period = time.Duration(quantity * float64(time.Second))
	case "milliseonds":
		period = time.Duration(quantity * float64(time.Millisecond))
	}
	return period, nil
}

// Shards - returns the number of shards to allocate this collection to
func (c *Config) Shards() int {
	if c.NumberOfShards == 0 {
		return 5
	}
	return c.NumberOfShards
}

// Replicas - returns the number of replicas to allocate this collection to
func (c *Config) Replicas() int {
	return c.NumberOfReplicas
}

// Schema - extract schema config
func (c *Config) Schema() (*Schema, error) {
	var ok bool
	for key, field := range c.SchemaCfg {
		if field.Type == "" {
			field.Type = String
			c.SchemaCfg[key] = field
		}
		if _, ok = FieldTypes[field.Type]; !ok {
			return nil, ErrUnsupportedType
		}
		if field.Length < 0 {
			return nil, ErrLengthLowerThanZero
		}
		if field.MaxLength < 0 {
			return nil, ErrLengthLowerThanZero
		}
		if field.Type == DateTime {
			if field.DateFormat == "" {
				field.DateFormat = time.RFC3339Nano
				c.SchemaCfg[key] = field
			}
			if _, ok = dateFormats[field.DateFormat]; !ok {
				return nil, ErrUnsupportedDateFormat
			}
		}
	}
	return &Schema{
		Fields: c.SchemaCfg,
	}, nil
}
