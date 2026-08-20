# search API

Base URL: `http://localhost:9012` (or `SEARCH_HOST`)

This service has no write endpoints. Everything in its index arrived via
the event stream -- see [../architecture.md](../architecture.md).

## `GET /ping` / `GET /info`

Same as every other service.

## `GET /search?q=&limit=&offset=`

Full-text search over active, non-removed, non-suspended-author posts,
ranked by Postgres `ts_rank`. `q` is required (`400` if empty or missing).
`limit` defaults to 20, capped at 100.

```json
{
  "status": "success",
  "data": {
    "results": [
      {
        "postId": "...",
        "authorId": "...",
        "authorUsername": "alice",
        "content": "hello world",
        "createdAt": "2026-08-20T20:35:11Z",
        "rank": 0.06
      }
    ]
  }
}
```
