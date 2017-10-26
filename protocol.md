# Protocol

> toq is a task queue based on Redis, it provide both producer and consumer writen in Golang, but the protocol is
generic purpose, which means both producer and consumer can write in other programming language, like Python, C,
Haskell, etc.

The only key in toq-protocol is it pass messages through broker in a generic way, producer generate a list of arguments
even keyword arguments, and pass it in a dumped json with a type of string. And consumer receive messages through
broker, and sent it to real worker, the worker decode json itself, and finally, produce a json, sent it back to toq(
by means, consumer), toq store the result in result backend.

Here is the workflow:

```
+--------------+    +------+    +-------------------------+    +--------+    +--------------+    +-------------+    +-----+
| toq producer | -> | JSON | -> | <dumped json in string> | -> | broker | -> | toq consumer | -> | result JSON | -> | ... |
+--------------+    +------+    +-------------------------+    +--------+    +--------------+    +-------------+    +-----+

```

And below is the interface toq must have(I wrote it in golang interface):

Task:

```go
type Task struct {
	ID         string   `json:"id"`
	Retry      bool     `json:"retry"` // retry or not, by default not
	MaxRetries int      `json:"max_retries"`
	Road       []string `json:"road"`  // all the queues the task enqueued by it's lifetime
	State      int      `json:"state"` // current task state
	Key        string   `json:"key"`   // the key to match which the function consumer runs
	Args       string   `json:"args"`  // dumped json string
}
```

Producer:

```go
type Enququer interface {
	Enqueue(queue string, t task.Task) error
}
```

Consumer:

```go
type Dequeuer interface {
	Dequeue() error
}
```
