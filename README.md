# distributed-social

A small event-driven social network backend in Go: three independently
deployable services that never share a database and never call each other
synchronously, coordinated entirely through an event stream.

- **accounts** — owns users and posts. The source of truth.
- **admins** — moderation. Suspends users, restores them, removes posts.
- **search** — a read-optimized, full-text-searchable view of every active
  post, built entirely by replaying events. It never queries `accounts`'
  database directly.

When `accounts` creates a post, it publishes a `post.created` event.
`search` picks that up and indexes it. When `admins` suspends a user, it
publishes `user.suspended`; both `accounts` (to mark the user suspended)
and `search` (to hide their posts from results) react to it independently.
No service reaches into another service's database — the event log is the
only integration point, which is what lets each of the three be built,
deployed, and scaled on its own.

## Quickstart

```bash
docker compose up -d --build
```

This starts NATS (JetStream enabled), one Postgres instance per service,
and all three services. Each service applies its own SQL migrations on
startup.

```bash
# create a user
curl -X POST localhost:9010/users -d '{"username":"alice","displayName":"Alice"}'

# create a post (use the id returned above)
curl -X POST localhost:9010/users/<user-id>/posts -d '{"content":"hello"}'

# give the event a moment to propagate, then search for it
curl "localhost:9012/search?q=hello"

# moderate: suspend the user, then confirm their posts vanish from search
curl -X POST localhost:9011/moderation/users/<user-id>/suspend \
  -d '{"actorId":"admin-1","reason":"policy violation"}'
curl "localhost:9012/search?q=hello"   # now empty
```

See [docs/architecture.md](docs/architecture.md) for how the pieces fit
together and [docs/api/](docs/api) for the full endpoint list per service.

## Local development without Docker

```bash
make up                # starts nats + the three postgres instances only
make run-dev-accounts   # in one terminal
make run-dev-admins     # in another
make run-dev-search     # in another
```

## Testing

```bash
make test
```

Domain logic (validation, event handling, cross-service consistency rules)
is unit tested against fake repositories and publishers — see
`internal/*/service_test.go`. There's no mocking of Postgres or NATS
themselves; those are exercised by actually running the stack, which is
also how a nil-pointer bug in the shared HTTP error-response path was
caught during development (see `utils/httputils/writer_test.go` for the
regression test).

## Tech stack

Go 1.25, `go.uber.org/fx` for dependency injection, `gorilla/mux`,
`jackc/pgx/v5` against Postgres (one instance per service, full-text
search via `tsvector`/`websearch_to_tsquery` in `search`), `nats-io/nats.go`
JetStream as the event bus, `zap`/`ipfs/go-log` for structured logging.

## Design notes and known trade-offs

- **At-least-once delivery, idempotent consumers.** Events can be
  redelivered; every consumer is written to be safe to run twice (upserts
  by natural key, status updates that just set a value rather than
  toggle it).
- **No transactional outbox.** `accounts` persists a post, then publishes
  its event as a second step. If the publish fails, the post exists but
  hasn't been indexed yet — it just waits for whatever event catches
  `search` up next, rather than the request failing outright. A
  production version of this would use a transactional outbox table to
  close that gap.
- **Each service owns its schema.** `accounts`, `admins`, and `search`
  each get their own Postgres instance in `docker-compose.yml`
  deliberately, not as one database with three schemas, so nothing can
  accidentally take a dependency on cross-service joins.
