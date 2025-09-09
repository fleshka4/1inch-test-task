# Uniswap V2 Off-chain Estimator

This project is a backend service that implements a single REST API endpoint for estimating swap amounts on Uniswap V2 pools.
It follows Clean Architecture principles, separating concerns into transport (HTTP),
service (business logic), infra (Ethereum client), and pure math layers. This makes the service easy to extend (e.g. add gRPC transport).

## Endpoints
### estimate

```shell
GET /estimate
```

Accepts query parameters:
- pool — address of Uniswap V2 pair contract
- src — address of source token
- dst — address of destination token
- src_amount — amount of source token (integer, respecting token decimals)

The response returns as a plain text integer in the smallest token units: the estimated `dst_amount`,
calculated off-chain using the reserves from the pool contract.

Example of usage:
```shell
curl "http://localhost:1337/estimate?pool=0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852&src=0xdAC17F958D2ee523a2206206994597C13D831ec7&dst=0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2&src_amount=10000000"
# => 6241000000000000
```

### ping

```shell
GET /estimate
```

Just easy healthcheck
```shell
curl http://localhost:1337/ping
# => pong
```

## Requirements
- Go 1.24.6+
- Ethereum RPC endpoint (Infura, Alchemy, QuickNode, or self-hosted)

## Configuration
- By default, the app expects `config/config.yaml`.
- If missing, you must create it or copy from `config/config.yaml.example`, do not forget to replace value in `rpc_url`.
- Alternatively, you can set `CONFIG_PATH` env variable to specify a custom config.

## Build
Command to build project:
```shell
make build
```
Command to build binary:
```shell
make build-server
```
Generate mocks:
```shell
make mocks
```

## Run linter
```shell
make lint
```

## Run tests
```shell
make tests
```
Run with coverage
```shell
make coverage
```
Run benchmark
```shell
make benchmark
```

## Run server
```shell
make run
```
or
```shell
go run ./cmd/server/main.go
```
