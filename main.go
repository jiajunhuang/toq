package main

import (
	"flag"
	"fmt"

	"github.com/jiajunhuang/toq/producer"
	"github.com/jiajunhuang/toq/task"
	"github.com/sirupsen/logrus"
)

func main() {
	flag.Parse()

	redisPool := NewRedisPool()
	p := producer.NewProducer(redisPool)
	for {
		logrus.Println("enqueue a job")
		t := task.Task{ID: fmt.Sprintf("task_%d", UUID4()), Key: "test_key", Args: "{}"}
		p.Enqueue("test_toq_queue", t)
	}
}
