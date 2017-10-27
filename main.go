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
