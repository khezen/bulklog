
# *bulklog*

Collects, buffers, and outputs logs across multiple sources and destinations.

![icon.png](https://github.com/khezen/bulklog/raw/master/.doc/icon.png)

*bulklog* is written in go and requires little resource.

*bulklog* supports memory and [redis](https://redis.io/) buffering.
*bulklog* also supports failover and can be set up for high availability.

---

## Concepts

*bulklog* tries to structure data as JSON since it has enough structure to be accessible while providing felxibility.

### Collection

A collection is a set of declarative informations about how *bulklog* should process data.

### Output

*bulklog* outputs JSON docuemts to destinations such as Elasticsearch, MongoDB, etc...

---

## Install

### Docker

[![khezen/bulklog](https://images.microbadger.com/badges/image/docker.pkg.github.com/khezen/bulklog/bulklog.svg)](https://github.com/khezen/bulklog/packages)

```bash
docker run -p 5017:5017 -v /etc/bulklog:/etc/bulklog docker.pkg.github.com/khezen/bulklog/bulklog:stable
```

#### Supported tags

* `latest`
* `2.0.0`, `2.0`, `2`, `stable`
* `1.0.10`, `1.0`, `1`


#### ENV

| key | Description | Default Value|
|---|---|---|
|CONFIG_PATH|path to the configuration folder|/etc/bulklog|

### Kubernetes

#### Helm

Deploy bulklog to a kubernetes cluster using Helm.

```bash
helm repo add khezen https://khezen.github.com/charts
helm install docker.pkg.github.com/khezen/bulklog/bulklog --name bulklog
```

---

## Config

Default [config.yaml](https://github.com/khezen/bulklog/raw/master/config.yaml).

### Persistence

Peristence is disabled by default in which case data is buffered in memory.
If enabled, it uses Redis(>= 2.4) to persist documents buffer. 
[Learn how to tune Redis persistence](https://redis.io/topics/persistence) for your requirements. 

```yaml
persistence:
  enabled: true
  redis:
    endpoint: localhost:6379
    password: changeme #(optional)
    db: 0 #(optional, default:0)
    idle_conn: 2 #(optional, default: 0)
    max_conn: 10 #(optional, defaut: no limit)
```

### Output

provides declarative information about *bulklog* output.

```yaml
output:
  elasticsearch:
    enabled: true
    endpoint: localhost:9200
    scheme: http
#   aws_auth:
#     access_key_id: changeme
#     secret_access_key: changeme
#     region: eu-west-1
#   basic_auth:
#     username: elastic
#     password: changeme
```

*from version 2.0.0 bulklog supports Elasticsearch 7.0.0 and above*

### Collections

examples:

```yaml
collections:
  - name: logs
    flush_period: 5 seconds # hours|minutes|seconds|milliseconds
    retention_period: 45 minutes
    schema: {}
```

*bulklog* is schema free but we encourage you to provide some base structure since it might enbale output destination to process data more efficiently.

```yaml
collections:
  - name: logs
    flush_period: 5 seconds # hours|minutes|seconds|milliseconds
    retention_period: 45 minutes
    shards: 6
    schema:
      source: 
        type: string
        max_length: 64
      stream: 
        type: string
        length: 6
      event: 
        type: string
      time:
        type: datetime
        date_format: 2006-01-02T15:04:05.999999999Z07:00
```

Even in the case above, *bulklog* remains schema free enabling log decoration with additional field.

#### collection

* **name**: `{collection name}`
* **flush_period**: `{duration}`
  * flush buffer to output every `{duration}`
* **retention_period**: `{duration}`
  * if an output is unavailable, **retention_period** set how long *bulklog* tries to output data to this output
  * if the output is unavailable for too long, **retention_period** ensure that *bulklog* will not accumulate too much data and will be able to serve other outputs.
* **shards**: the number of shards to allocate this index to. Check [Elasticsearch documentation](https://www.elastic.co/guide/en/elasticsearch/guide/2.x/scale.html) to learn more about it.
* **schema**: `{map of fields by field name}`

#### field

* **type**: `{field type}`
  * see [supported types](#supported-types)
* **length**: `{field exact length}` (optional,string only)
* **max_length**: `{field maximum length}` (optional, string only)
* **date_format**: `{date time formatting}` (optional, datetime only)

---

## API

### push document

```http
POST bulklog/v1/{collectionName} HTTP/1.1
Content-Type: application/json
{
  ...
}

HTTP/1.1 200 OK
```

example:

```http
POST bulklog/v1/logs HTTP/1.1
Content-Type: application/json
{
  "source":"service1",
  "stream": "stderr",
  "event": "divizion by zero",
  "time": "2018-11-15T14:12:12Z"
}

### push documents in batches

```http
POST /v1/{collectionName}/batch HTTP/1.1
Content-Type: application/json
{...}
{...}

HTTP/1.1 200 OK
```

example:

```http
POST bulklog/v1/logs/batch HTTP/1.1
Content-Type: application/json
{"source":"service1","stream": "stderr","event": "divizion by zero","time" : "2019-01-13T19:30:12"}
{"source":"service1","stream": "stdout","event": "successfully processed","time" : "2019-01-13T19:35:12"}

HTTP/1.1 200 OK
```

### health

```http
GET bulklog/liveness HTTP/1.1

HTTP/1.1 200 OK
```

```http
GET bulklog/readiness HTTP/1.1

HTTP/1.1 200 OK
```

---

## supported types

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
  * *bulklog* doesn't check the date format. Most outputs accept any even if it defers from the configured one
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

## Issues

If you have any problems or questions, please ask for help through a [GitHub issue](https://github.com/khezen/bulklog/issues).

## Contributions

Help is always welcome! For example, documentation (like the text you are reading now) can always use improvement. There's always code that can be improved. If you ever see something you think should be fixed, you should own it. If you have no idea what to start on, you can browse the issues labeled with [help wanted](https://github.com/khezen/bulklog/labels/help%20wanted).

As a potential contributor, your changes and ideas are welcome at any hour of the day or night, weekdays, weekends, and holidays. Please do not ever hesitate to ask a question or send a pull request.

[Code of conduct](https://github.com/khezen/bulklog/blob/master/CODE_OF_CONDUCT.md).
