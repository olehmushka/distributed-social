# admins API

Base URL: `http://localhost:9011` (or `ADMINS_HOST`)

## `GET /ping` / `GET /info`

Same as every other service.

## `POST /moderation/users/{id}/suspend`

```json
{"actorId": "admin-1", "reason": "policy violation"}
```

Records a `suspend_user` moderation action against this service's own
audit log and publishes `admins.user.suspended`. `accounts` and `search`
react to that independently -- this endpoint does not call either of
them. `actorId` and `reason` are both required (`400` if missing).

## `POST /moderation/users/{id}/restore`

Same shape as suspend; publishes `admins.user.restored`.

## `POST /moderation/posts/{id}/remove`

```json
{"actorId": "admin-1", "reason": "abusive content"}
```

Publishes `admins.post.removed`.

## `GET /moderation/actions?limit=&offset=`

Returns the audit log, most recent first. `limit` defaults to 20, capped
at 100.
