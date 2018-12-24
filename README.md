
# *bulklog*

Collects, buffers, and outputs logs across multiple sources and destinations.

![](https://github.com/khezen/bulklog/raw/master/icon.png)

*bulklog* is written in go and requires little resource.

*bulklog* supports memory and [redis](https://redis.io/topics/persistence) buffering to prevent data loss. 
*bulklog* also supports failover and can be set up for high availability.

---

## Concepts

*bulklog* tries to structure data as JSON since it has enough structure to be accessible while providing felxibility.

### Schema

A Schema provides declarative informations about how *bulklog* should process data.

### Collection

A collection is a set of schemas.

### Consumer

*bulklog* outputs JSON docuemts to consumers such as Elasticsearch, MongoDB, etc...

---

## quick start

### Docker
`docker run -p 5000:5000 -v /etc/bulklog:/etc/bulklog khezen/bulklog:stable`
#### Supported tags
* `1`
* `2.0.0`, `1.0`, `1`, `stable`, `latest`

---

## Config 

Default [config.yaml](https://github.com/khezen/bulklog/raw/master/config.yaml).

### Redis

Redis is disabled by default in which case data is buffered in memory.

```yaml
redis:
  enabled: true
  address: http://localhost:6379
  password: changeme
  db: 0
```

### Consumers

```yaml
consumers: 
 
  elasticsearch:
    enabled: true
    addr: http://localhost:9200
    shards: 1
#   aws_auth:
#     access_key_id: changeme
#     secret_access_key: changeme
#     region: eu-west-1
#   basic_auth: 
#     username: elastic
#     password: changeme
```

### Collections

* collection
  * name: {collection name}
  * flush_period: {duration}
    * flush buffer and output data to consumers every {duration} 
  * retention_period: {duration}
    * if a consumer is unavailable, retention_period set how long *bulklog* tries to output data to this consumer
    * if the consumer is unavailable for too long, retention_period ensure that *bulklog* will not accumulate too much data and will be able to serve other consumers.
  * schemas: {map of schema configurations by schema name}

* schema

```yaml
collections:

- name: web
  flush_period: 5 seconds # hours|minutes|seconds|milliseconds
  retention_period: 45 minutes

  schemas:
    trace:
      source: 
        type: string
      request_id: 
        type: string
        length: 36
      client_ip: 
        type: string
      host: 
        type: string
      path: 
        type: string
      method: 
        type: string
        max_length: 6
      request_dump: 
        type: string
      status_code: 
        type: int16
      response_dump: 
        type: string
      response_time_ms: 
        type: int32
```

# User Feedback
## Issues
If you have any problems or questions, please ask for help through a [GitHub issue](https://github.com/khezen/bulklog/issues).
