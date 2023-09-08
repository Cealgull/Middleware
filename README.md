# Cealgull Middleware

![Unittests](https://github.com/Cealgull/Verify/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/Cealgull/Middleware/graph/badge.svg?token=BGKUR08BRW)](https://codecov.io/gh/Cealgull/Middleware)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

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

### Smoke Test

`tests/py/smoke` contains python scripts to pursue the smoke test for endpoints correctness. To run the smoke test suite,

first set the constant in `tests/py/smoke/config.py` for two servers. Then run the following command,

```console
user@localhost:/path/to/middleware/tests/py/ $ pipx install cryptography

user@localhost:/path/to/middleware/tests/py/ $ python -m smoke
```

### Load Test

'tests/py/locust' this is the load tests based on the famous [locust](https://locust.io/) framework. To Install and run,

``` console
user@localhost:/path/to/middleware/tests/py/ $ pipx install locust

user@localhost:/path/to/middleware/tests/py/locust $ locust --users 25 --host http://localhost:8089 -f locustfile.py
```

