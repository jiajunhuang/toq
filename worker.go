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
