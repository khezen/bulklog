
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

## Quick Start

### Docker

```bash
docker run -p 5000:5000 -v /etc/bulklog:/etc/bulklog khezen/bulklog:stable
```

#### Supported tags

* `latest`
* `1.0.3`, `1.0`, `1`, `stable`

---

## Config 

Default [config.yaml](https://github.com/khezen/bulklog/raw/master/config.yaml).

### Persistence

Redis is disabled by default in which case data is buffered in memory.

```yaml
redis:
  enabled: true
  address: http://localhost:6379
  password: changeme
  db: 0
```

### Output

provides declarative information about consumers which *bulklog* output data to.

```yaml
output: 
 
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

example:

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

#### collection

* **name**: `{collection name}`
* **flush_period**: `{duration}`
  * flush buffer and output data to consumers every `{duration}` 
* **retention_period**: `{duration}`
  * if a consumer is unavailable, **retention_period** set how long *bulklog* tries to output data to this consumer
  * if the consumer is unavailable for too long, **retention_period** ensure that *bulklog* will not accumulate too much data and will be able to serve other consumers.
* **schemas**: `{map of schema configurations by schema name}`

#### schema

map of fields by field name

#### field

* **type**: `{field type}`
* **length**: `{field exact length}` (optional,string only)
* **max_length**: `{field maximum length}` (optional, string only)
* **date_format**: `{date time formatting}` (optional, datetime only)


##### supported types:

* **bool** : `True` or `False`

* **unint8** : `0` to `255`
* **uint16** : `0` to `65535`
* **uint32** : `0` to `4294967295`
* **unit64** : `0` to `18446744073709551615`

* **int8** : `-128` to `127`
* **int16** : `-32768` to `32767`
* **int32** : `-2147483648` to `2147483647`
* **int64** : `-9223372036854775808` to `9223372036854775807`

* **float32** : `-3.40282346638528859811704183484516925440e+38` to `3.40282346638528859811704183484516925440e+38`
* **float64** : `-1.797693134862315708145274237317043567981e+308` to `1.797693134862315708145274237317043567981e+308`

* **string** : sequence of characters
   * **lenght**: string exact length
   * **max_length**: string maximum length

* **datetime** : `1970-01-01T00:00:00.000000000Z` (example)
  * *bulklog* doesn't check the date format. Most consumers accept any even if it defers from the configured one
  * **date_format**: date format string
    * `Mon Jan _2 15:04:05 2006`
    * `Mon Jan _2 15:04:05 MST 2006`
    * `Mon Jan 02 15:04:05 -0700 2006`
    * `02 Jan 06 15:04 MST`
    * `02 Jan 06 15:04 -0700`
    * `Monday, 02-Jan-06 15:04:05 MST`
    * `Mon, 02 Jan 2006 15:04:05 MST`
    * `Mon, 02 Jan 2006 15:04:05 -0700`
    * `2006-01-02T15:04:05Z07:00`
    * `2006-01-02T15:04:05.999999999Z07:00` (**default**)
    * `3:04PM`
    * `Jan _2 15:04:05`
    * `Jan _2 15:04:05.000`
    * `Jan _2 15:04:05.000000`
    * `Jan _2 15:04:05.000000000`
    * `2006-01-02 15:04:05 MST`
    * `2006-01-02 15:04:05.999999999 MST`

* **object** : inner document 

---

## API

### Collection

```http
POST /bulklog/{collectionName}/{schemaName} HTTP/1.1
Content-Type: application/json
{
  ...
}

HTTP/1.1 200 OK
```

example:

```http
POST /bulklog/logs/log HTTP/1.1
Content-Type: application/json
{
  "source":"service1",
  "request_id":"cd603a72-f74c-4f2c-afeb-bc29f788db78",
  "level": "Fatal",
  "message": "divizion by zero"
}

HTTP/1.1 200 OK
```

### Health

```http
GET /bulklog/health HTTP/1.1

HTTP/1.1 200 OK
```

---

## Issues
If you have any problems or questions, please ask for help through a [GitHub issue](https://github.com/khezen/bulklog/issues).

## Contributions

Help is always welcome! For example, documentation (like the text you are reading now) can always use improvement. There's always code that can be improved. If you ever see something you think should be fixed, you should own it. If you have no idea what to start on, you can browse the issues labeled with [help wanted](https://github.com/khezen/bulklog/labels/help%20wanted).

As a potential contributor, your changes and ideas are welcome at any hour of the day or night, weekdays, weekends, and holidays. Please do not ever hesitate to ask a question or send a pull request.
