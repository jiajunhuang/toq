package task

const (
	ResultStateFailed = iota
	ResultStateSucceed
)

type Result struct {
	TaskID     string `json:"task_id"`
	WannaRetry bool   `json:"wanna_retry"` // if it wanna retry
	State      int    `json:"state"`       // succeed or failed
	Message    string `json:"message"`
	Result     string `json:"result"` // dumped json it want return
}
