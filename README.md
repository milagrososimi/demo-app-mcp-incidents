# payments platform

Three services that together take an order from queue to authorized payment.

| Service | Responsibility |
| --- | --- |
| `payments-api` | Serves payment lookups over HTTP |
| `order-consumer` | Drains the order queue and hands work downstream |
| `payment-service` | Authorizes orders against the upstream processor |

Everything is standard library. There is nothing to install and no broker or
database to run: the queue and the processor are stood up in-process so a
service can be started and watched on its own.

## Running

```bash
go build ./...
go vet ./...
go test ./...

LISTEN_ADDR=:8080 go run ./cmd/payments-api    # curl localhost:8080/payments/p1
go run ./cmd/order-consumer                     # drains a backlog, reports throughput
go run ./cmd/payment-service                    # authorizes against a local stub
```

Each service logs its version on startup. Builds stamp it from the release
tag; an unstamped build reports `dev`.

```bash
go build -ldflags "-X main.version=$(git describe --tags)" ./cmd/order-consumer
```

At debug level the consumer also reports, per batch, how long the poll round
trip took and how long the orders in it took to handle.

## Configuration

Every setting is read from the environment. The manifests under `deploy/` are
what supply them in each cluster, so the same variable can be set locally to
reproduce what a deployment does.

### payments-api

| Variable | Default | |
| --- | --- | --- |
| `LISTEN_ADDR` | `:8080` | Address to serve on |
| `DB_POOL_SIZE` | `50` | Concurrent database connections |
| `DB_QUERY_DURATION` | `20ms` | How long one lookup holds its connection |
| `DB_ACQUIRE_TIMEOUT` | `2s` | How long a request queues for a connection |

### order-consumer

| Variable | Default | |
| --- | --- | --- |
| `CONSUMER_BATCH_SIZE` | `500` | Orders fetched per poll |
| `CONSUMER_POLL_COST` | `40ms` | Round trip to the broker |
| `CONSUMER_WORK_COST` | `100us` | Handling cost per order |
| `CONSUMER_BACKLOG` | `5000` | Orders waiting at startup |

### payment-service

| Variable | Default | |
| --- | --- | --- |
| `UPSTREAM_URL` | local stub | Processor base URL |
| `UPSTREAM_TIMEOUT` | `5s` | Per attempt, not per authorization |
| `UPSTREAM_RETRIES` | `2` | Extra attempts after the first |
| `ORDER_ID` | `ord_1001` | Order to authorize on startup |

## Maintenance

The history here is referenced from outside the repository. Please do not
rewrite it.
