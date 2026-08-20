// Package eventsapi is the shared event contract between services: subject
// names and payload shapes that producers and consumers both depend on,
// analogous to a schema registry in a larger deployment.
package eventsapi

import "time"

const (
	SubjectPostCreated   = "accounts.post.created"
	SubjectPostRemoved   = "admins.post.removed"
	SubjectUserSuspended = "admins.user.suspended"
	SubjectUserRestored  = "admins.user.restored"
)

// StreamSubjects lists every subject the shared JetStream stream must accept.
var StreamSubjects = []string{"accounts.>", "admins.>"}

type PostCreatedPayload struct {
	PostID         string    `json:"postId"`
	AuthorID       string    `json:"authorId"`
	AuthorUsername string    `json:"authorUsername"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
}

type PostRemovedPayload struct {
	PostID string `json:"postId"`
}

type UserSuspendedPayload struct {
	UserID string `json:"userId"`
}

type UserRestoredPayload struct {
	UserID string `json:"userId"`
}
