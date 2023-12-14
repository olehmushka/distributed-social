package schemas

type Status string

const (
	SuccessStatus Status = "success"
	FailureStatus Status = "failure"
	ErrorStatus   Status = "error"
)

type SuccessResp[T interface{}] struct {
	Status   Status    `json:"status"`
	Data     T         `json:"data"`
	Metadata *Metadata `json:"metadata"`
}

type FailureResp[T interface{}] struct {
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
