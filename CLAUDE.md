# distributed-social

See [AGENTS.md](AGENTS.md) for the full project orientation (layout,
hard rules, verification steps) -- it applies here unchanged. This file
only adds Claude-Code-specific notes.

## Claude-Code-specific notes

- Project permissions live in `.claude/settings.json` (build/test/vet/
  docker/read-only git commands are pre-allowed).
- Use `internal/accounts/service_test.go` as the template when asked to
  add domain logic to `admins` or `search` -- same fake-repository,
  fake-publisher pattern.
- If asked to touch anything under `api/*/module.go`, re-read
  [docs/architecture.md](docs/architecture.md) first: the fx wiring
  there (pool → repository → service → handlers, plus the separate
  consumer registration) is the pattern to extend, not something to
  restructure without a reason.
