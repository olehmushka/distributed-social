// Package search is a read-only, full-text-searchable view of active
// posts. It has no write API of its own -- every row in its index
// exists because an event from accounts or admins put it there, which
// means the whole index can be rebuilt by replaying the event stream.
package search

import "time"

type Document struct {
	PostID         string
	AuthorID       string
	AuthorUsername string
	Content        string
	CreatedAt      time.Time
}

type Result struct {
	Document
	Rank float64
}
