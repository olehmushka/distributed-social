// Package admins implements moderation: suspending/restoring users and
// removing posts. It keeps its own audit log and publishes an event for
// every action; it never writes to accounts' or search's data directly.
package admins

import "time"

type TargetType string

const (
	TargetUser TargetType = "user"
	TargetPost TargetType = "post"
)

type ActionType string

const (
	ActionSuspendUser ActionType = "suspend_user"
	ActionRestoreUser ActionType = "restore_user"
	ActionRemovePost  ActionType = "remove_post"
)

type ModerationAction struct {
	ID        string
	ActorID   string
	Target    TargetType
	TargetID  string
	Action    ActionType
	Reason    string
	CreatedAt time.Time
}
