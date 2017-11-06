# toq

toq pronounces `to queue`, which is a simple distributed task queue based on [Redis](https://redis.io), it is inspired
by [Python-RQ](http://python-rq.org/).

## Usage

toq is a realy simple task queue, if you want to implement a worker, you just have to define a function whose type is:

```go
type Worker func(task.Task) task.Result
```

And register it with a key to consumer engine(of course you have to initialize a consumer engine first), and call
`c.Dequeue()`, then everything is just ok, consumer engine will auto discover tasks and match the specify worker by
the key you given. just like this:

```go
package main

import (
	"flag"
	"time"

	"github.com/jiajunhuang/toq/consumer"
	"github.com/jiajunhuang/toq/task"
	"github.com/sirupsen/logrus"
)

var (
	redisPasswd = flag.String("redisPasswd", "", "")
	redisURI    = flag.String("redisURI", "", "")
	redisDBNum  = flag.Int("redisDBNum", 0, "")
	maxIdle     = flag.Int("maxIdle", 1024, "")
	maxActive   = flag.Int("maxActive", 100, "")
)

func Run(t task.Task) task.Result {
	logrus.Infof("running task %s...", t.ID)
	time.Sleep(10 * time.Second)
	logrus.Infof("task %s succeed", t.ID)

	return task.Result{TaskID: t.ID, State: task.ResultStateSucceed, Message: "succeed"}
}

func main() {
	flag.Parse()

	redisPool := NewRedisPool()
	c := consumer.NewConsumer(redisPool, []string{"test_toq_queue"})
	if err := c.RegisterWorker("test_key", Run); err != nil {
		logrus.Errorf("failed to register worker with error: %s", err)
	}
	c.Dequeue()
}
```

Really simple, right?

> for now, toq just support redis as broker, but it will support more in the future.

## Producer

If you want to produce tasks, you just have to do something like this:

```go
package main

import (
	"flag"

	"github.com/jiajunhuang/toq/producer"
	"github.com/jiajunhuang/toq/task"
	"github.com/sirupsen/logrus"
)

func main() {
	flag.Parse()

	redisPool := NewRedisPool()
	p := producer.NewProducer(redisPool)
	for {
		taskID := UUID4()
		logrus.Infof("enqueue task %s", taskID)
		t := task.Task{ID: taskID, Key: "test_key", Args: "{}"}
		p.Enqueue("test_toq_queue", t)
	}
}
```

If you want to learn more, please visit https://github.com/jiajunhuang/toq

## Features

- [x] Concurrent task executor
- [x] Easy to learn and use
- [x] Language non-specific protocol
- [x] Distributed(with support by broker, like Redis)
- [x] Retry
- [x] Graceful restart
- [ ] Exception/Panic handle
- [ ] Retry countdown
- [ ] HA
