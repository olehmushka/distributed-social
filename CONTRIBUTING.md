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

## Git conventions

### Branching

GitHub Flow: `main` is always deployable. There are no long-lived
`develop`/`release` branches.

1. Branch off `main`.
2. Open a PR back into `main` once it's ready for review.
3. Merge (squash preferred, so `main`'s history stays one commit per
   logical change) and delete the branch.

Branch names are `<type>/<kebab-case-description>`, using the same
`<type>` vocabulary as commits below:

```
feat/search-result-pagination
fix/nil-pointer-in-writer
docs/update-architecture-diagram
chore/bump-nats-go
```

### Commit messages

[Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[(scope)]: <short summary, imperative mood, no trailing period>

<body -- explain why, not what; the diff already shows what changed>
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`,
`chore`, `ci`, `build`. **Scope** is optional and, in this repo, is
usually a service name (`accounts`, `admins`, `search`) or an area
(`docs`, `deps`). A breaking change gets a `!` after the type/scope
(`feat(search)!: ...`) or a `BREAKING CHANGE:` footer.

```
feat(search): add pagination to /search

Large result sets were slow to render with everything on one page.
Adds limit/offset params with the same defaults as the other list
endpoints.

fix(accounts): stop double-counting post length

chore(deps): bump nats.go to v1.54.0
```

If a PR is squash-merged, its title becomes the commit message on
`main` -- follow the same format there.

### Before opening a PR

```bash
make fmt lint test
docker compose up -d --build   # and exercise the change against the real stack
```

CI (`.github/workflows/ci.yml`) runs build, vet, gofmt check,
race-enabled tests, and a Docker build on every PR; all four must pass.
Unit tests run against fakes; they will not catch a bug in how a
handler talks to Postgres, NATS, or the shared HTTP response writer.
This project has already shipped one nil-pointer panic that only
showed up under live traffic (see `utils/httputils/writer_test.go`) --
don't rely on mocked tests alone for anything that touches the HTTP or
event-bus boundary.

## Reporting bugs / requesting features

Use the issue templates under `.github/ISSUE_TEMPLATE/`. For security
vulnerabilities, see [SECURITY.md](SECURITY.md) instead of opening a
public issue.
