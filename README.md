# Harvester [![CircleCI](https://circleci.com/gh/thebeatapp/harvester.svg?style=svg)](https://circleci.com/gh/thebeatapp/harvester) [![codecov](https://codecov.io/gh/thebeatapp/harvester/branch/master/graph/badge.svg)](https://codecov.io/gh/thebeatapp/harvester) [![Go Report Card](https://goreportcard.com/badge/github.com/thebeatapp/harvester)](https://goreportcard.com/report/github.com/thebeatapp/harvester) [![GoDoc](https://godoc.org/github.com/thebeatapp/harvester?status.svg)](https://godoc.org/github.com/thebeatapp/harvester) ![GitHub release](https://img.shields.io/github/release/thebeatapp/harvester.svg)

`Harvester` is a configuration library which helps setting up and monitoring configuration values in order to dynamically
reconfigure your application.

Configuration can be obtained from the following sources:

- Seed values, are hard-coded values into your configuration struct
- Environment values, are obtained from the environment
- Consul, which is used to get initial values and to monitor them for changes

The order is applied as it is listed above. Consul seeder and monitor are optional and will be used only if `Harvester` is created with the above components.

`Harvester` expects a go structure with tags which defines one or more of the above like the following:

```go
type Config struct {
    Name    sync.String  `seed:"John Doe"`
    Age     sync.Int64   `seed:"18" env:"ENV_AGE"`
    IsAdmin sync.Bool    `seed:"true" env:"ENV_IS_ADMIN" consul:"/config/is-admin"`
}
```

The above defines the following fields:

- Name, which will be seeded with the value `John Doe`
- Age, which will be seeded with the value `18`, and if exists, overridden with whatever value the env var `ENV_AGE` holds
- IsAdmin, which will be seeded with the value `true`, and if exists, overridden with whatever value the env var `ENV_AGE` holds and then from consul if the consul seeder and/or watcher are provided.

The fields have to be on of the types that the sync package supports in order to allow concurrent read and write to the fields. The following types are supported:

- sync.String, allows for concurrent string manipulation
- sync.Int64, allows for concurrent int64 manipulation
- sync.Float64, allows for concurrent float64 manipulation
- sync.Bool, allows for concurrent bool manipulation

`Harvester` has a seeding phase and an optional monitoring phase.

## Seeding phase
  
- Apply the seed tag value, if present
- Apply the value contained in the env var, if present
- Apply the value returned from Consul, if present and harvester is setup to seed from consul

Conditions where seeding fails:

- If at the end of the seeding phase one or more fields have not been seeded
- If the seed value is invalid

### Seeder

`Harvester` allows the creation of custom getters which are used by the seeder and implement the following interface:

```go
type Getter interface {
    Get(key string) (string, error)
}
```

Seed and env tags are supported by default, the Consul getter has to be setup when creating a `Harvester` with the builder.

## Monitoring phase (Consul only)
  
- Monitor a key and apply if tag key matches
- Monitor a key-prefix and apply if tag key matches

### Monitor

`Harvester` allows for dynamically changing the config value by monitoring a source. The following sources are available:

- Consul, which supports monitoring for keys and key-prefixes.

This feature have to be setup when creating a `Harvester` with the builder.

## Builder

The `Harvester` builder pattern is used to create a `Harvester` instance. The builder supports setting up:

- Consul seed, for setting up seeding from Consul
- Consul monitor, for setting up monitoring from Consul

```go
    h, err := New(cfg).
                WithConsulSeed("address", "dc", "token").
                WithConsulMonitor("address", "dc", "token", items...).
                Create()
```

The above snippet set's up a `Harvester` instance with consul seed and monitor.

## Consul

Consul has support for versioning (`ModifyIndex`) which allows us to change the value only if the version is higher than the one currently applied.

## Todo

Will be github issues once we move to thebeatapp organization.

- The code has a data race and cannot be used with safety.
- The config struct should check for null fields and initialize with default pointer value
- create examples folder with a service implementation (@d.baltas)
- move to circle-ci
- Error handling
  - Logging
  - return chan error and let the client handle it
- Support change events which can be fired (chan with change...)
- Support nesting (structs...)

## How to Contribute

See [Contribution Guidelines](CONTRIBUTE.md).

## Code of conduct

Please note that this project is released with a [Contributor Code of Conduct](https://www.contributor-covenant.org/adopters). By participating in this project and its community you agree to abide by those terms.