# Examples

## Prerequisites

For examples `02_consul` and `03_consul_monitor` we need Consul obviously.
A fast way to get consul is the following:

    wget https://releases.hashicorp.com/consul/1.4.3/consul_1.4.3_linux_amd64.zip  
    unzip "consul_1.4.3_linux_amd64.zip"
    ./consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul -dev -bind=$(hostname -I | awk '{print $1}' | xargs) -http-port 8500

## 01 Basic usage with env vars

    go run examples/01_basic/main.go

    2019/09/19 11:18:36 INFO: field IndexName updated with value customers-v1, version: 0
    2019/09/19 11:18:36 INFO: seed value customers-v1 applied on field IndexName
    2019/09/19 11:18:36 INFO: field CacheRetention updated with value 43200, version: 0
    2019/09/19 11:18:36 INFO: seed value 43200 applied on field CacheRetention
    2019/09/19 11:18:36 INFO: field CacheRetention updated with value 86400, version: 0
    2019/09/19 11:18:36 INFO: env var value 86400 applied on field CacheRetention
    2019/09/19 11:18:36 INFO: field LogLevel updated with value DEBUG, version: 0
    2019/09/19 11:18:36 INFO: seed value DEBUG applied on field LogLevel
    2019/09/19 11:18:36 WARN: flag var loglevel did not exist for field LogLevel
    2019/09/19 11:18:36 Config : IndexName: customers-v1, CacheRetention: 86400, LogLevel: DEBUG

## 02 Seed values from Consul

    go run examples/02_consul/main.go

    2019/09/19 11:27:29 INFO: field IndexName updated with value customers-v1, version: 0
    2019/09/19 11:27:29 INFO: seed value customers-v1 applied on field IndexName
    2019/09/19 11:27:29 INFO: field CacheRetention updated with value 43200, version: 0
    2019/09/19 11:27:29 INFO: seed value 43200 applied on field CacheRetention
    2019/09/19 11:27:29 INFO: field CacheRetention updated with value 86400, version: 0
    2019/09/19 11:27:29 INFO: env var value 86400 applied on field CacheRetention
    2019/09/19 11:27:29 INFO: field LogLevel updated with value DEBUG, version: 0
    2019/09/19 11:27:29 INFO: seed value DEBUG applied on field LogLevel
    2019/09/19 11:27:29 INFO: field OpeningBalance updated with value 0.0, version: 0
    2019/09/19 11:27:29 INFO: seed value 0.0 applied on field OpeningBalance
    2019/09/19 11:27:29 WARN: env var ENV_CONSUL_VAR did not exist for field OpeningBalance
    2019/09/19 11:27:29 INFO: field OpeningBalance updated with value 100.0, version: 12
    2019/09/19 11:27:29 INFO: consul value 100.0 applied on field OpeningBalance
    2019/09/19 11:27:29 WARN: flag var loglevel did not exist for field LogLevel
    2019/09/19 11:27:29 Config: IndexName: customers-v1, CacheRetention: 86400, LogLevel: DEBUG, OpeningBalance: 100.000000

## 03 Monitor Consul for live changes

    go run examples/03_consul_monitor/main.go

    2019/09/19 11:31:13 INFO: field IndexName updated with value customers-v1, version: 0
    2019/09/19 11:31:13 INFO: seed value customers-v1 applied on field IndexName
    2019/09/19 11:31:13 INFO: field CacheRetention updated with value 43200, version: 0
    2019/09/19 11:31:13 INFO: seed value 43200 applied on field CacheRetention
    2019/09/19 11:31:13 INFO: field CacheRetention updated with value 86400, version: 0
    2019/09/19 11:31:13 INFO: env var value 86400 applied on field CacheRetention
    2019/09/19 11:31:13 INFO: field LogLevel updated with value DEBUG, version: 0
    2019/09/19 11:31:13 INFO: seed value DEBUG applied on field LogLevel
    2019/09/19 11:31:13 INFO: field OpeningBalance updated with value 0.0, version: 0
    2019/09/19 11:31:13 INFO: seed value 0.0 applied on field OpeningBalance
    2019/09/19 11:31:13 WARN: env var ENV_CONSUL_VAR did not exist for field OpeningBalance
    2019/09/19 11:31:13 INFO: field OpeningBalance updated with value 100.0, version: 31
    2019/09/19 11:31:13 INFO: consul value 100.0 applied on field OpeningBalance
    2019/09/19 11:31:13 WARN: flag var loglevel did not exist for field LogLevel
    2019/09/19 11:31:13 INFO: plan for key harvester/example_03/openingbalance created
    2019/09/19 11:31:13 Config: IndexName: customers-v1, CacheRetention: 86400, LogLevel: DEBUG, OpeningBalance: 100.000000
    2019/09/19 11:31:13 WARN: version 31 is older or same as the field's OpeningBalance
    2019/09/19 11:31:14 INFO: field OpeningBalance updated with value 999.99, version: 33
    2019/09/19 11:31:15 Config: IndexName: customers-v1, CacheRetention: 86400, LogLevel: DEBUG, OpeningBalance: 999.990000
