package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/jiajunhuang/toq/producer"
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

func main() {
	flag.Parse()

	redisPool := NewRedisPool()
	p := producer.NewProducer(redisPool)
	for {
		logrus.Println("enqueue a job")
		t := task.Task{ID: fmt.Sprintf("task_%d", UUID4()), Key: "test_key", Args: "{}"}
		p.Enqueue("test_toq_queue", t)
		time.Sleep(1 * time.Second)
	}
}
