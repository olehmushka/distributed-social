// Package schemas defines the shared HTTP response envelope every
// service's handlers write: SuccessResp, FailureResp, and ErrorResp,
// each carrying the same Metadata block (request id, duration, version).
package schemas

type Status string

const (
	SuccessStatus Status = "success"
	FailureStatus Status = "failure"
	ErrorStatus   Status = "error"
)

type SuccessResp[T any] struct {
	Status   Status    `json:"status"`
	Data     T         `json:"data"`
	Metadata *Metadata `json:"metadata"`
}

type FailureResp[T any] struct {
	Status   Status    `json:"status"`
	Data     T         `json:"data"`
	Message  string    `json:"message"`
	Metadata *Metadata `json:"metadata"`
}

type ErrorResp struct {
	Status   Status    `json:"status"`
	Message  string    `json:"message"`
	Metadata *Metadata `json:"metadata"`
}

type Metadata struct {
	RequestID string `json:"requestId"`
	Duration  int    `json:"duration"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}
