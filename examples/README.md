# Examples

## 01 Basic usage with env vars
```
12:41:20 ➜  harvester git:(examples) ✗  go run examples/01_basic/main.go
2019/05/26 12:42:40 Attribute State: name: John Doe, age: 18
12:42:40 ➜  harvester git:(examples) ✗  ENV_AGE=33 go run examples/01_basic/main.go
2019/05/26 12:42:53 Attribute State: name: John Doe, age: 33
```
## 02 Seed values from Consul
```
12:42:53 ➜  harvester git:(examples) ✗  go run examples/02_consul/main.go
2019/05/26 12:43:45 Attribute State: name: John Doe, age: 18, consulVar: boo
```
## 03 Monitor Consul for live changes
```
12:43:45 ➜  harvester git:(examples) ✗  go run examples/03_consul_monitor/main.go
2019/05/26 12:44:14 Attribute State: name: John Doe, age: 18, consulVar: bar
# after changing the consul value
2019/05/26 12:44:34 Attribute State: name: John Doe, age: 18, consulVar: baz
```
