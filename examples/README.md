# Examples

## Prerequisites

For examples `02_consul` and `03_consul_monitor` we need Consul obviously.
A fast way to get consul is the following:

    wget https://releases.hashicorp.com/consul/1.4.3/consul_1.4.3_linux_amd64.zip  
    unzip "consul_1.4.3_linux_amd64.zip"
    ./consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul -dev -bind=$(hostname -I | awk '{print $1}' | xargs) -http-port 8500

For example `04_vault`, it is possible to run Vault with Docker Compose (https://docs.docker.com/compose/install/) using
the file `examples/04_vault/docker-compose.yml`:

    cd examples/04_vault
    docker-compose up -d
    go run main.go

## 01 Basic usage with env vars

    go run examples/01_basic/main.go

    2019/06/04 22:07:59 INFO: seed value John Doe applied on Name
    2019/06/04 22:07:59 INFO: seed value 18 applied on Age
    2019/06/04 22:07:59 INFO: env var value 25 applied on Age
    2019/06/04 22:07:59 Config : Name: John Doe, Age: 25

## 02 Seed values from Consul

    go run examples/02_consul/main.go

    2019/07/12 14:40:04 INFO: field Name updated with value John Doe, version: 0
    2019/07/12 14:40:04 INFO: seed value John Doe applied on field Name
    2019/07/12 14:40:04 INFO: field Age updated with value 18, version: 0
    2019/07/12 14:40:04 INFO: seed value 18 applied on field Age
    2019/07/12 14:40:04 INFO: field Age updated with value 25, version: 0
    2019/07/12 14:40:04 INFO: env var value 25 applied on field Age
    2019/07/12 14:40:04 INFO: field City updated with value London, version: 0
    2019/07/12 14:40:04 INFO: seed value London applied on field City
    2019/07/12 14:40:04 INFO: field Balance updated with value 99.9, version: 0
    2019/07/12 14:40:04 INFO: seed value 99.9 applied on field Balance
    2019/07/12 14:40:04 WARN: env var ENV_CONSUL_VAR did not exist for field Balance
    2019/07/12 14:40:04 INFO: field Balance updated with value 123.45, version: 7642
    2019/07/12 14:40:04 INFO: consul value 123.45 applied on field Balance
    2019/07/12 14:40:04 WARN: flag var city did not exist for field City
    2019/07/12 14:40:04 Config: Name: John Doe, Age: 25, City: London, Balance: 123.450000

## 03 Monitor Consul for live changes

    go run examples/03_consul_monitor/main.go

    2019/07/12 14:41:30 INFO: field Name updated with value John Doe, version: 0
    2019/07/12 14:41:30 INFO: seed value John Doe applied on field Name
    2019/07/12 14:41:30 INFO: field Age updated with value 18, version: 0
    2019/07/12 14:41:30 INFO: seed value 18 applied on field Age
    2019/07/12 14:41:30 INFO: field Age updated with value 25, version: 0
    2019/07/12 14:41:30 INFO: env var value 25 applied on field Age
    2019/07/12 14:41:30 INFO: field City updated with value London, version: 0
    2019/07/12 14:41:30 INFO: seed value London applied on field City
    2019/07/12 14:41:30 INFO: field Balance updated with value 99.9, version: 0
    2019/07/12 14:41:30 INFO: seed value 99.9 applied on field Balance
    2019/07/12 14:41:30 INFO: field Balance updated with value 123.45, version: 7647
    2019/07/12 14:41:30 INFO: consul value 123.45 applied on field Balance
    2019/07/12 14:41:30 WARN: flag var city did not exist for field City
    2019/07/12 14:41:30 INFO: plan for key harvester/example_03/balance created
    2019/07/12 14:41:30 Config: Name: John Doe, Age: 25, City: London, Balance: 123.450000
    2019/07/12 14:41:30 WARN: version 7647 is older or same as the field's Balance
    2019/07/12 14:41:31 INFO: field Balance updated with value 999.99, version: 7649
    2019/07/12 14:41:32 Config: Name: John Doe, Age: 25, City: London, Balance: 999.990000

## 04 Seed values from Vault

    go run examples/04_vault/main.go

    2019/07/18 13:45:45 INFO: field Secret updated with value @pps3cr3t, version: 0
    2019/07/18 13:45:45 INFO: vault value @pps3cr3t applied on field Secret
    2019/07/18 13:45:45 Secret: @pps3cr3t