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
