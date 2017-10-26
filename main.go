package main

import (
	"flag"
	"fmt"

	"github.com/garyburd/redigo/redis"
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

func NewRedisPool() *redis.Pool {
	uri := fmt.Sprintf("redis://:%s@%s/%d", *redisPasswd, *redisURI, *redisDBNum)
	return &redis.Pool{
		MaxIdle:   *maxIdle,
		MaxActive: *maxActive,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(uri)
			if err != nil {
				logrus.Panicf("connect to redis(%s) got error: %s", *redisURI, err)
			}
			return c, nil
		},
	}
}

func main() {
	flag.Parse()

	redisPool := NewRedisPool()
	p := producer.NewProducer(redisPool)
	t := task.Task{ID: "0", Key: "toq_worker", Args: "{}"}
	p.Enqueue("test_toq_queue", t)
}
