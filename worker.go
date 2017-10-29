package main

import (
	"flag"
	"time"

	"github.com/jiajunhuang/toq/consumer"
	"github.com/jiajunhuang/toq/task"
	"github.com/sirupsen/logrus"
)

func Run(t task.Task) task.Result {
	logrus.Infof("running task %s...", t.ID)
	time.Sleep(10 * time.Second)

	return task.Result{TaskID: t.ID, WannaRetry: true, State: task.ResultStateFailed, Message: "failed"}
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
