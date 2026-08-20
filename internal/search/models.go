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
