## What and why

<!-- What does this change do, and why is it needed? Link an issue if there is one. -->

## How it was tested

<!-- go test ./... covers logic against fakes. For anything touching HTTP
handlers, Postgres, or the NATS event bus, also describe how you
exercised it against the real stack (`docker compose up -d --build`),
since that boundary isn't covered by unit tests alone. -->

## Checklist

- [ ] `make fmt lint test` passes
- [ ] If this touches a service boundary (an event subject, a response
      shape, cross-service behavior), `docs/architecture.md` or the
      relevant `docs/api/*.md` is updated
- [ ] If this adds an event consumer, it's idempotent under redelivery
