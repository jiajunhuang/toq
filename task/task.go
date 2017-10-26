package task

type Task struct {
	ID         string   `json:"id"`
	Retry      bool     `json:"retry"` // retry or not, by default not
	MaxRetries int      `json:"max_retries"`
	Road       []string `json:"road"`  // all the queues the task enqueued by it's lifetime
	State      int      `json:"state"` // current task state
	Key        string   `json:"key"`   // the key to match which the function consumer runs
	Args       string   `json:"args"`  // dumped json string
}
