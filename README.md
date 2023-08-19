# Cealgull Middleware

Cealgull Middleware is a middleware which receive requests from Cealgull app and interop with Hyperledger Fabric, PostgresDB and a simple Kubo/IPFS Gateway.


## Getting Started

You may need a `config.yaml` file in the root directory of the project. `config_template.yaml` provides you

with details about how to connect to a fabric-samples test-network and a kubo container.

## Testing

### Unit Test

To Run unit tests, you need [Mockery](https://vektra.github.io/mockery/latest/installation/). To install, simply run

```console
user@localhost:~ $ go install github.com/vektra/mockery/v2@v2.32.4
user@localhost:~ $ mockery
user@localhost:~ $ go test ./...
```

### Load Test

Cealgull middleware provides two sets of test suite, `tests/k6` contains major [k6](https://k6.io/) load testing setup and 

'tests/py' contains smoke tests and cryptography-driven tests based on [locust](https://locust.io/). To Install, 

``` console
user@localhost:~ $ pipx install locust
user@localhost:~ $ go install go.k6.io/xk6/cmd/xk6@latest
user@localhost:~ $ xk6 build --with github.com/szkiba/xk6-dashboard@latest --with github.com/grafana/xk6-output-influxdb
user@loclahost:tests/k6 $ yarn install && yarn build
```

