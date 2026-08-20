package schemas

import "time"

type ModerationActionRespData struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actorId"`
	TargetType string    `json:"targetType"`
	TargetID   string    `json:"targetId"`
	Action     string    `json:"action"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ListModerationActionsRespData struct {
	Actions []ModerationActionRespData `json:"actions"`
}
