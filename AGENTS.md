# AGENTS.md

Instructions for AI coding agents working in this repository. (Claude
Code specifically also reads [CLAUDE.md](CLAUDE.md), kept in sync with
this file.)

## What this is

Event-driven Go microservices: `accounts` (users/posts), `admins`
(moderation), `search` (full-text index). Each owns its own Postgres
instance; the three only ever talk to each other through NATS
JetStream events. Read [docs/architecture.md](docs/architecture.md)
before changing anything that crosses a service boundary -- the event
contract in `internal/eventsapi` is the API between services, not any
HTTP call.

## Layout

- `cmd/<service>/` -- entrypoint (`urfave/cli` flags, `fx.New(...)`).
- `api/<service>/` -- HTTP layer: handlers, routes, fx module wiring.
- `internal/<service>/` -- domain logic: models, `Repository` interface,
  Postgres implementation, `Service`, unit tests, embedded migrations.
- `internal/eventbus/` -- NATS JetStream wrapper.
- `internal/eventsapi/` -- shared event contract (subjects + payloads).
- `internal/pg/` -- shared Postgres connection + migration runner.
- `schemas/` -- HTTP response envelope types.
- `openapi/` -- machine-readable API specs per service.
- `docs/` -- architecture and per-service API documentation.

## Hard rules

1. **No service reaches into another service's database or calls its
   HTTP API to read/write data.** Cross-service coordination is a new
   subject in `internal/eventsapi`, published by one service and
   consumed by whichever others need it -- never a direct call.
2. **Every event consumer must be idempotent.** Delivery is
   at-least-once. Upsert by natural key; set absolute status values
   rather than toggling.
3. **Don't reintroduce dead utility code.** `utils/` was previously a
   dumping ground for unused helpers (13 unused slice functions, an
   unused stats-based random generator pulling in `gonum`) that were
   deleted. Only add something to `utils/` once code outside `utils/`
   actually calls it.
4. **Watch the `err`-shadowing bug class.** `utils/httputils/writer.go`
   shipped a real nil-pointer panic once: `metadata, err :=
   GetMetadata(ctx)` inside a function that already had an `err`
   parameter silently clobbered it. `CreateFail`/`CreateError` name that
   parameter `inErr` specifically to avoid recreating this -- keep that
   convention if you touch this file.
5. **Branch and commit naming follow [CONTRIBUTING.md](CONTRIBUTING.md#git-conventions):**
   branches as `<type>/<kebab-description>`, commits as Conventional
   Commits (`feat:`, `fix:`, `docs:`, ...). If you're committing on
   behalf of a user, use that format unless they say otherwise.

## Verifying a change

```bash
go build ./... && go vet ./... && gofmt -l .   # must be clean/empty
go test ./...                                   # unit tests, no docker needed
docker compose up -d --build                    # then exercise it for real
```

Unit tests in this repo run against fake repositories/publishers (see
`internal/accounts/service_test.go`) -- they do not exercise Postgres,
NATS, or the shared HTTP writer. The nil-pointer bug above was only
caught by testing against the real running stack; don't consider an
HTTP-handler or event-consumer change done until you've done the same
(`docker compose up -d --build`, then `curl` the endpoint or trigger the
event and confirm the behavior, not just that it compiles).
