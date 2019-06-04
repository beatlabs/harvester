# Examples

## Prerequisites

For examples `02_consul` and `03_consul_monitor` we need Consul obviously.
A fast way to get consul is the following:

    wget https://releases.hashicorp.com/consul/1.4.3/consul_1.4.3_linux_amd64.zip  
    unzip "consul_1.4.3_linux_amd64.zip"
    ./consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul -dev -bind=$(hostname -I | awk '{print $1}' | xargs) -http-port 8500

## 01 Basic usage with env vars

    go run examples/01_basic/main.go

    2019/06/04 22:07:59 INFO: seed value John Doe applied on Name
    2019/06/04 22:07:59 INFO: seed value 18 applied on Age
    2019/06/04 22:07:59 INFO: env var value 25 applied on Age
    2019/06/04 22:07:59 Config : Name: John Doe, Age: 25

## 02 Seed values from Consul

    go run examples/02_consul/main.go

    2019/06/04 22:05:04 INFO: seed value John Doe applied on Name
    2019/06/04 22:05:04 INFO: seed value 18 applied on Age
    2019/06/04 22:05:04 WARN: env var ENV_AGE did not exist for Age
    2019/06/04 22:05:04 INFO: seed value 99.9 applied on Balance
    2019/06/04 22:05:04 WARN: env var ENV_CONSUL_VAR did not exist for Balance
    2019/06/04 22:05:04 INFO: consul value 123.45 applied on Balance
    2019/06/04 22:05:04 Config: Name: John Doe, Age: 18, Balance: 123.450000

## 03 Monitor Consul for live changes

    go run examples/03_consul_monitor/main.go

    2019/06/04 22:05:32 INFO: seed value John Doe applied on Name
    2019/06/04 22:05:32 INFO: seed value 18 applied on Age
    2019/06/04 22:05:32 WARN: env var ENV_AGE did not exist for Age
    2019/06/04 22:05:32 INFO: seed value 99.9 applied on Balance
    2019/06/04 22:05:32 WARN: env var ENV_CONSUL_VAR did not exist for Balance
    2019/06/04 22:05:32 INFO: consul value 123.45 applied on Balance
    2019/06/04 22:05:32 Config: Name: John Doe, Age: 18, Balance: 123.450000
    2019/06/04 22:05:34 Config: Name: John Doe, Age: 18, Balance: 999.990000
