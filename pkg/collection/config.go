package collection

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Config -
type Config struct {
	Name               Name                        `yaml:"name"`
	FlushPeriodStr     string                      `yaml:"flush_period"`
	RetentionPeriodStr string                      `yaml:"retention_period"`
	SchemasCfg         map[SchemaName]SchemaConfig `yaml:"schemas"`
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
		break
	case "minutes":
		period = time.Duration(quantity * float64(time.Minute))
		break
	case "seconds":
		period = time.Duration(quantity * float64(time.Second))
		break
	case "milliseonds":
		period = time.Duration(quantity * float64(time.Millisecond))
		break
	}
	return period, nil
}

// Schemas - extract schemas config
func (c *Config) Schemas() ([]Schema, error) {
	schemas := make([]Schema, 0, len(c.SchemasCfg))
	for schemaName, fields := range c.SchemasCfg {
		var ok bool
		for key, field := range fields {
			if field.Type == "" {
				field.Type = String
				fields[key] = field
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
				}
				if _, ok = dateFormats[field.DateFormat]; !ok {
					return nil, ErrUnsupportedDateFormat
				}
			}
		}
		schemas = append(schemas, Schema{
			Name:   schemaName,
			Fields: fields,
		})
	}
	return schemas, nil
}
