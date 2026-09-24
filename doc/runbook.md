# Runbook

## Which service does what

| Service | Shape | Where it runs from |
| --- | --- | --- |
| `payments-api` | long-running HTTP server | `deploy/payments-api.yaml` |
| `order-consumer` | drains the queue and exits | `deploy/order-consumer.yaml` |
| `payment-service` | authorizes one order and exits | `deploy/payment-service.yaml` |

## Where settings come from

Every setting is an environment variable, and the manifests under `deploy/`
are what set them. Nothing is read from a file at runtime, so a change in
behaviour is a change to a manifest or to the code that reads it — there is no
third place to look.

Code defaults apply when a variable is unset. A service started locally with no
environment therefore behaves as the code says, not as the cluster does. Use
the manifest values when reproducing something seen in production.

## Reading the logs

Each service logs its version on startup, then one line per unit of work. The
version is stamped from the release tag at build time; an unstamped build
reports `dev`.

An unparseable setting is reported and the default is used rather than
refusing to start — a service that will not come up cannot say what was wrong
with its configuration.

## Who to page

Whoever is on call for the payments platform. There is no separate rotation per
service.
