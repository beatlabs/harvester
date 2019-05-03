# Harvester configuration package

`Harvester` is a configuration library which helps setting up and monitoring configuration values in order to dynamically 
reconfigure your application.

Configuration can be obtained from the following sources:

- Seed values, are hard-coded values into your configuration struct
- Environment values, are obtained from the environment
- Consul, which is used to get initial values and to monitor them for changes

The order is applied as it is listed above.

`Harvester` expects a go structure with tags which defines one or more of the above like the following:

```go
type Config struct {
    Name    string  `seed:"John Doe"`
    Age     int64   `seed:"18" env:"ENV_AGE"`
    Balance float64 `seed:"99.9" env:"ENV_BALANCE"`
    HasJob  bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
}
```

## How to Contribute

See [Contribution Guidelines](CONTRIBUTE.md).

## Code of conduct

Please note that this project is released with a [Contributor Code of Conduct](https://www.contributor-covenant.org/adopters). By participating in this project and its community you agree to abide by those terms.

## Todo

- Support nesting (structs...)
- consul client timeouts should be set
- OPTIONAL: Support change events which can be fired (chan with change...)
- move to circle-ci
- Error handling
  - Logging
  - return chan error and let the client handle it
