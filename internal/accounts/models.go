// Package accounts is the source of truth for users and posts. It's the
// only package in this repo allowed to write to the accounts database,
// and it publishes events (see internal/eventsapi) after every write so
// other services can react without ever reading this data directly.
package accounts

import "time"

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
)

type PostStatus string

const (
	PostStatusActive  PostStatus = "active"
	PostStatusRemoved PostStatus = "removed"
)

type User struct {
	ID          string
	Username    string
	DisplayName string
	Status      UserStatus
	CreatedAt   time.Time
}

type Post struct {
	ID        string
	AuthorID  string
	Content   string
	Status    PostStatus
	CreatedAt time.Time
}
