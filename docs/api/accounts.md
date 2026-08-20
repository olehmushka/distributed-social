# accounts API

Base URL: `http://localhost:9010` (or `ACCOUNTS_HOST`)

All responses are wrapped in the shared envelope (`schemas.SuccessResp` /
`FailureResp` / `ErrorResp`): `{"status": "...", "data": ..., "metadata": {...}}`.

## `GET /ping`

Liveness check. `{"ok": true}`.

## `GET /info`

`{"name": "accounts"}`.

## `POST /users`

```json
{"username": "alice", "displayName": "Alice"}
```

Username must be 3-32 characters of letters, digits, or underscores.
Returns `409` if the username is already taken, `400` on validation
failure. On success, `201`-equivalent (`200` with `status: "success"`)
with the created user.

## `GET /users/{id}`

Returns the user, or `404` if it doesn't exist.

## `POST /users/{id}/posts`

```json
{"content": "hello world"}
```

Content must be non-empty and at most 500 characters. Returns `403` if
the author is currently suspended, `404` if the author doesn't exist,
`400` on validation failure. On success, publishes `accounts.post.created`
to the event stream (see [../architecture.md](../architecture.md)).

## `GET /posts/{id}`

Returns the post, or `404`.

## `GET /users/{id}/posts?limit=&offset=`

Lists a user's posts, most recent first. `limit` defaults to 20, capped
at 100.
