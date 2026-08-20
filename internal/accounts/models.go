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
