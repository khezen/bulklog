package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/khezen/bulklog/collection"
)

// Index - elasticsearch index definition
// ref: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-templates.html
type Index struct {
	Template string        `json:"template"`
	Settings IndexSettings `json:"settings"`
	Mappings Mappings      `json:"mappings"`
}

// IndexSettings -
type IndexSettings struct {
	NumberOfShards int `json:"number_of_shards"`
}

// Mappings - document schema definitions
type Mappings map[collection.SchemaName]Mapping

// Mapping - document schema definition
// ref : // ref: https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping.html
type Mapping struct {
	Properties map[string]Field `json:"properties"`
}

// Field -
type Field struct {
	Type string `json:"type"`
}

// RenderElasticIndex - render elasticsearch mapping
func RenderElasticIndex(collect collection.Collection, settings IndexSettings) Index {
	index := Index{
		Template: fmt.Sprintf("%s-*", collect.Name),
		Settings: settings,
		Mappings: make(map[collection.SchemaName]Mapping),
	}
	for _, schema := range collect.Schemas {
		mapping := Mapping{
			Properties: make(map[string]Field),
		}
		for key, field := range schema.Fields {
			mapping.Properties[key] = Field{
				Type: translateType(field),
			}
		}
		index.Mappings[schema.Name] = mapping
	}
	return index
}

func translateType(field collection.Field) string {
	switch field.Type {
	case collection.Bool:
		return "bool"
	case collection.UInt8, collection.UInt16, collection.UInt32, collection.UInt64,
		collection.Int8, collection.Int16, collection.Int32, collection.Int64:
		return "long"
	case collection.Float32, collection.Float64:
		return "double"
	case collection.DateTime:
		return "time"
	case collection.Object:
		return "object"
	case collection.String:
		if field.MaxLength > 0 || field.Length > 0 {
			return "keyword"
		}
		return "text"
	default:
		return "text"
	}
}

// RenderIndexName - logs: logs-2017.05.26
func RenderIndexName(d collection.Document) string {
	indexBuf := bytes.NewBufferString(string(d.CollectionName))
	indexBuf.WriteString("-")
	indexBuf.WriteString(d.PostedAt.Format("2006.01.02"))
	return indexBuf.String()
}

// Digest returns the JSON request to be append to the bulk
func Digest(d collection.Document) ([]byte, error) {
	request := make(map[string]interface{})
	//{ "index" : { "_index" : "logs-2017.05.28", "_type" : "log", "_id" : "1" } }
	docDescription := make(map[string]interface{})
	docDescription["_index"] = RenderIndexName(d)
	docDescription["_type"] = d.SchemaName
	docDescription["_id"] = d.ID
	docDescription["post_date"] = d.PostedAt.Format(time.RFC3339)
	request["index"] = docDescription
	body, err := json.Marshal(request)
	body = append(body, '\n')
	body = append(body, d.Body...)
	body = append(body, '\n')
	if err != nil {
		return nil, err
	}
	return body, nil
}
