# Architecture

## Services and data ownership

```
                       ┌─────────────┐
        writes          │             │
   ┌────────────────────▶  accounts   │◀───────────────┐
   │                    │  (users,    │                 │
   │                    │   posts)    │                 │ user.suspended
   │                    └──────┬──────┘                 │ user.restored
   │                           │ post.created            │
   │                           ▼                         │
   │                    ┌─────────────┐          ┌───────┴──────┐
   │        post.removed│             │          │              │
   └────────────────────┤    NATS     │◀─────────┤    admins    │
                         │  JetStream  │ post.removed  (moderation│
                         │  (events)   │          │   audit log) │
                         └──────┬──────┘          └──────────────┘
                                │ post.created
                                │ post.removed
                                │ user.suspended / user.restored
                                ▼
                         ┌─────────────┐
                         │   search    │
                         │ (read-only  │
                         │  index)     │
                         └─────────────┘
```

Each box has its own Postgres database. No service holds a connection
string to another service's database, and no service calls another
service's HTTP API to read or write its data. The only channel between
them is the event stream.

- **accounts** is the only service that can create a user or a post.
  Every write happens against its own Postgres instance, and only after
  that write commits does it publish an event describing what happened.
- **admins** never touches `accounts`' or `search`'s tables. A moderation
  action is: validate the request, write an audit-log row to its own
  database, publish an event. That's it. It has no way to force a post to
  disappear except by publishing `post.removed` and letting `accounts`
  and `search` react to it on their own schedule.
- **search** has no write API at all. Every row in its `documents` table
  exists because some event put it there. Deleting `search`'s database
  and replaying the event stream from the beginning would fully rebuild
  it -- that's the point of a read-optimized view built from events
  rather than a service with its own independent state.

## Why this shape

The interesting engineering problem in a system like this isn't the CRUD
-- it's what happens when a piece of information needs to be true in more
than one place. A post's "is this visible" status genuinely lives in two
places: `accounts` (the record itself) and `search` (the index entry).
Two ways to keep those in sync were considered:

1. Have `admins` call `accounts`' API directly to mark a post removed, and
   `accounts` call `search`'s API to update the index. This is simpler to
   trace through, but it means `admins` can't do its job unless `accounts`
   is up, `accounts` can't finish an update unless `search` is up, and
   scaling or redeploying any one service becomes a coordination problem
   for the others.
2. Have every service publish what happened to it, and let every
   interested service subscribe. `admins` doesn't need `accounts` or
   `search` to be reachable to record that a post was removed -- it just
   needs NATS. `accounts` and `search` catch up whenever they're next
   able to.

This project uses (2). The cost is real: consistency is eventual, not
immediate (there's a window between a suspension being recorded and it
being reflected in search results), and every consumer has to be written
to tolerate a message arriving twice. Both of those trade-offs are
visible in the code -- see `internal/search/service.go`'s idempotent
`Handle*` methods and the "Design notes" section of the [README](../README.md).

## Event contract

Subjects and payload shapes live in `internal/eventsapi`, imported by
whichever services produce or consume them. Nothing enforces schema
compatibility across services beyond that shared Go package -- in a
larger system this is where a schema registry with compatibility checking
would sit.

| Subject               | Published by | Consumed by      |
|------------------------|--------------|-------------------|
| `accounts.post.created` | accounts     | search            |
| `admins.post.removed`   | admins       | accounts, search  |
| `admins.user.suspended` | admins       | accounts, search  |
| `admins.user.restored`  | admins       | accounts, search  |

Delivery is at-least-once (JetStream, durable pull consumers, manual ack
only after the handler succeeds). Every consumer in this repo is written
so redelivery is harmless: `search`'s writes are upserts keyed by post id,
and `accounts`'/`search`'s status updates set an absolute value rather
than toggling one.
