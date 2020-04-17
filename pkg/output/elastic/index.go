package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/khezen/bulklog/pkg/collection"
)

// Index - elasticsearch index definition
// ref: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-templates.html
type Index struct {
	Pattern  string        `json:"index_patterns"`
	Settings IndexSettings `json:"settings"`
	Mappings Mapping       `json:"mappings"`
}

// IndexSettings -
type IndexSettings struct {
	NumberOfShards   int `json:"number_of_shards"`
	NumberOfReplicas int `json:"number_of_replicas"`
}

// Mapping - document schema definition
// ref : // ref: https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping.html
type Mapping struct {
	Properties map[string]Field `json:"properties"`
}

// Field -
type Field struct {
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
}

// RenderElasticIndex - render elasticsearch mapping
func RenderElasticIndex(collect *collection.Collection) Index {
	index := Index{
		Pattern: fmt.Sprintf("%s-*", collect.Name),
		Settings: IndexSettings{
			NumberOfShards:   collect.Shards,
			NumberOfReplicas: collect.Replicas,
		},
		Mappings: Mapping{
			Properties: make(map[string]Field),
		},
	}
	for key, field := range collect.Schema.Fields {
		index.Mappings.Properties[key] = Field{
			Type: translateType(field),
		}
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
		return "date"
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
	//{ "index" : { "_index" : "logs-2017.05.28", "_id" : "1" } }
	docDescription := make(map[string]interface{})
	docDescription["_index"] = RenderIndexName(d)
	docDescription["_id"] = d.ID
	request["index"] = docDescription
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal.%s", err)
	}
	body = append(body, '\n')
	body = append(body, d.Body...)
	body = append(body, '\n')
	return body, nil
}
