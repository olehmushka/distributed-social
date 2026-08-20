# distributed-social

Event-driven Go microservices: `accounts` (users/posts), `admins`
(moderation), `search` (full-text index). Each owns its own Postgres
instance and the three only ever talk to each other through NATS
JetStream events -- see [docs/architecture.md](docs/architecture.md)
before changing anything that crosses a service boundary.

## Layout

- `cmd/<service>/` -- entrypoint (`urfave/cli` flags, `fx.New(...)`).
- `api/<service>/` -- HTTP layer: handlers, routes, fx module wiring
  (DB pool, JetStream connection, repository, domain service, event
  consumers).
- `internal/<service>/` -- domain logic: models, `Repository` interface,
  Postgres implementation, `Service` (validation + business rules +
  event publishing/handling), unit tests against fake repos/publishers,
  embedded SQL migrations.
- `internal/eventbus/` -- NATS JetStream wrapper (envelope, publish,
  durable pull-consumer subscription helper).
- `internal/eventsapi/` -- the shared event contract (subjects + payload
  structs) every producer and consumer imports.
- `internal/pg/` -- shared Postgres connection + embedded-migration
  runner.
- `schemas/` -- HTTP response envelope types.
- `utils/` -- small shared helpers (httputils, contextutils, middlewares
  support). Keep this directory lean -- it was previously a dumping
  ground for unused code (13 unused slice helpers, an unused stats-based
  random float generator) that got deleted; don't let that regrow. Only
  add a utility here once something outside `utils/` actually calls it.

## Conventions

- New domain logic goes in `internal/<service>/service.go` with a fake
  `Repository`/`Publisher` in `service_test.go` -- see
  `internal/accounts/service_test.go` for the pattern. Don't reach for
  a real Postgres/NATS integration test for business-rule coverage.
- Every event consumer must be idempotent (upsert by natural key, or set
  an absolute status rather than toggling) -- JetStream delivery here is
  at-least-once, not exactly-once.
- A service never opens a connection to another service's database or
  calls its HTTP API to read/write data. If two services need to agree
  on something, that's a new subject in `internal/eventsapi`, not a new
  cross-service call.
- HTTP error responses go through `utils/httputils.Writer`
  (`WriteFail`=400, `WriteStatus`=explicit code, `WriteError`=500). Watch
  for the shadowing bug class that shipped once already: `metadata, err
  := GetMetadata(ctx)` inside a function that already has an `err`
  parameter silently overwrites it. `CreateFail`/`CreateError` name that
  parameter `inErr` specifically to avoid this -- keep that naming if you
  touch this file.

## Running things

```bash
docker compose up -d --build   # everything, including migrations on boot
make up                        # just nats + the three postgres instances
make run-dev-accounts          # then run a service natively against `make up`
make test                      # go test ./...
```

Unit tests never need Docker. If you change anything in the event flow
or the HTTP error paths, verify it against the real stack (`docker
compose up -d --build` + `curl`) before considering it done -- the writer
bug above was only caught that way; the mocked unit tests never exercised
it.
