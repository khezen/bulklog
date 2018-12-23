
# What is *espipe*?

*espipe* is a service indexing documents in bulk requests to Elasticsearch.

![](https://github.com/khezen/espipe/raw/master/espipe.png)



# Run

## Docker
`docker run -p 5000:5000 -v /etc/espipe:/etc/espipe khezen/espipe:stable`
### Supported tags
* `1`
* `2.0.0`, `1.0`, `1`, `stable`,`latest`

# Config
```json
{
  "redis": {},
  "elasticsearch": {}
}
```

## Redis
Redis is disabled by default in which case buffer is stored in memory.
```json
  "redis": {
    "enabled": true,
    "address":"http://localhost:6379",
    "password": "changeme",
    "parition":0
  }
```

## Elasticsearch
```json
  "elasticsearch": {
    "address":"http://localhost:9200",
    "templates": [{
      "name": "logs",
      "flushPeriodMS": 5000,
      "body": {...}
    }]
  }
```
### templates
  * *espipe* creates index templates in elasticsearch if they do not exist yet
  * for each template, *espipe* creates indices on daily basis
    * template=logs-\*,
    * indices=logs-2017.01.05, logs-2017.01.06, etc..
  * For each index, *espipe* triggers bulk requests every `flushPeriodMS`

#### template.body
template.body takes the template **settings** and **mappings** with types definition.
See the [Create Template API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-templates.html).

The mapping itself is flexible and is schema-free. New fields are automatically added to the type mapping definition when *espipe* indexes a new document. Check out the [mapping section](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping.html) for more information on mapping definitions.

### AWS Sign

*espipe* supports AWS authentication for Amazon Elasticsearch Service
```json

"elasticsearch": {
  "address":"https://host.eu-west-1.es.amazonaws.com",
  "AWSAuth": {
    "accessKeyId": "changeme",
    "secretAccessKey": "changeme",
    "region": "eu-west-1"
  },
  "templates": [...]
}
```

### Basic Auth

*espipe* supports basic authentication for Elasticsearch (shield, search-guard)

```json

"elasticsearch": {
  "address": "http://localhost:9200",
  "basicAuth":{ 
    "username": "elastic",
    "password": "changeme"
  },
  "templates": [...]
}
```

## Default config.json

* See [default config.json](https://github.com/khezen/espipe/raw/master/config.json).

* See [Go logger client](https://godoc.org/github.com/khezen/espipe/log) working with indexes defined in the [default config.json](https://github.com/khezen/espipe/raw/master/config.json).

Request|Response|Description
---|---|---
POST /espipe/v1/logs/log JSON body | 200 OK | indexes JSON body as `log` in Elasticsearch `logs-yyyy.MM.dd`
POST /espipe/v1/web/trace JSON body | 200 OK | indexes JSON body as `trace` in Elasticsearch `web-yyyy.MM.dd`
GET /espipe/v1/health | 200 OK | healthcheck





# User Feedback
## Issues
If you have any problems or questions, please ask for help through a [GitHub issue](https://github.com/khezen/espipe/issues).
