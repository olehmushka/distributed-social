# Contributing

## Getting set up

```bash
docker compose up -d --build   # nats, one postgres per service, all three services
make test                      # go test ./... -- no docker required
```

See [README.md](README.md) for the full quickstart and
[docs/architecture.md](docs/architecture.md) for how the services fit
together before making a change that crosses a service boundary.

## Ground rules for changes

- **A service never reaches into another service's database or calls its
  HTTP API to read/write data.** If two services need to agree on
  something, add a subject to `internal/eventsapi` and consume it --
  don't add a synchronous call between services.
- **Event consumers must be idempotent.** Delivery is at-least-once;
  redelivery of the same event must not corrupt state. Upsert by natural
  key, or set an absolute status rather than toggling it.
- **New domain logic gets a unit test against a fake repository and
  publisher**, not a live Postgres/NATS integration test -- see
  `internal/accounts/service_test.go` for the pattern.
- **Before opening a PR**, run:
  ```bash
  make fmt lint test
  docker compose up -d --build   # and exercise the change against the real stack
  ```
  Unit tests run against fakes; they will not catch a bug in how a
  handler talks to Postgres, NATS, or the shared HTTP response writer.
  This project has already shipped one nil-pointer panic that only
  showed up under live traffic (see `utils/httputils/writer_test.go`) --
  don't rely on mocked tests alone for anything that touches the HTTP or
  event-bus boundary.

## Commit messages and PRs

Explain *why*, not just *what* -- the diff already shows what changed.
Keep PRs scoped to one logical change. CI (`.github/workflows/ci.yml`)
runs build, vet, gofmt check, race-enabled tests, and a Docker build on
every PR; all four must pass.

## Reporting bugs / requesting features

Use the issue templates under `.github/ISSUE_TEMPLATE/`. For security
vulnerabilities, see [SECURITY.md](SECURITY.md) instead of opening a
public issue.
