# Changelog

All notable changes to this project are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this project
doesn't yet use tagged releases or semantic versioning.

## [Unreleased]

### Changed

- Rebuilt `accounts`, `admins`, and `search` from identical ping/info
  stubs into real services that own their own Postgres database and
  coordinate entirely through a shared NATS JetStream event stream. See
  [docs/architecture.md](docs/architecture.md).
- Rewrote git history to remove ~48MB of compiled binaries that had been
  committed directly, and to stop tracking `.env`.
- Removed unused utility code (`utils/sliceutils`, most of
  `utils/random`) and the `gonum`/`go-nanoid` dependencies that existed
  only to support it.

### Fixed

- Nil pointer panic in the shared HTTP error-response path
  (`utils/httputils.CreateFail`/`CreateError`), caused by a parameter
  being shadowed by an unrelated `:=` assignment.

### Added

- Unit tests for all new domain logic (`internal/*/service_test.go`),
  CI (build, vet, gofmt check, race-enabled tests, Docker build on every
  push/PR), and project documentation (README, `docs/architecture.md`,
  `docs/api/*.md`, OpenAPI specs, `CLAUDE.md`/`AGENTS.md`).
