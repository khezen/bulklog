
# What is *espipe*?

*espipe* is a service indexing documents in bulk requests to Elasticsearch.

![](https://github.com/khezen/espipe/raw/master/espipe.png)



# Run
`docker run -p 5000:5000 -v /etc/espipe:/etc/espipe khezen/espipe:6`
## Supported tags and respective `Dockerfile` links
* `6`, `latest`
 [(6/Dockerfile)](https://github.com/khezen/espipe/blob/6/Dockerfile)[![](https://images.microbadger.com/badges/image/khezen/espipe.svg)](https://hub.docker.com/r/khezen/espipe/)
# Services

Request|Response|Description
---|---|---
POST /espipe/{template.name}/{documentType}  JSON body | 200 OK | indexes JSON body as `{documentType}` in Elasticsearch `{template.name}-yyyy.MM.dd`
GET /espipe/health | 200 OK | healthcheck

# Configure
```json
{
    "elasticsearch": "http://localhost:9200",
    "templates": [{
            "name": "logs",
            "bufferSizeKB": 5000,
            "timerMS": 5000,
            "body": {...}
        }
    ]
}

```
## Configure templates

* templates is an array of template configurations:
  * *espipe* creates templates in elasticsearch if they do not exist yet
  * an index template will automatically be applied when new indices are created
  * for each template, *espipe* creates indices on daily basis
    * **example:**
      * template=logs-\*,
      * indices=logs-2017.01.05, logs-2017.01.06, etc..


* *espipe* indexes documents in bulk requests. For each index, *espipe* triggers bulk requests when:
  * bulk size >= template.bufferSizeKB
  * ticker event, period=template.timerMS

### template.body
template.body takes the template **settings** and **mappings** with types definition.
See the [Create Template API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-templates.html).

The mapping itself is flexible and is schema-free. New fields are automatically added to the type mapping definition when *espipe* indexes a new document. Check out the [mapping section](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping.html) for more information on mapping definitions.

## Security

### AWS Sign

*espipe* supports AWS authentication for Amazon Elasticsearch Service
```json
{
  "elasticsearch": "https://host.eu-west-1.es.amazonaws.com",
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
{
    "elasticsearch": "http://localhost:9200",
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
POST /espipe/logs/log JSON body | 200 OK | indexes JSON body as `log` in Elasticsearch `logs-yyyy.MM.dd`
POST /espipe/web/trace JSON body | 200 OK | indexes JSON body as `trace` in Elasticsearch `web-yyyy.MM.dd`
GET /espipe/health | 200 OK | healthcheck





# User Feedback
## Issues
If you have any problems or questions, please ask for help through a [GitHub issue](https://github.com/khezen/espipe/issues).
