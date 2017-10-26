package consumer

import (
	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
)

type Dequeuer interface {
	Dequeue() error
}

type Consumer struct {
	redisPool *redis.Pool
	queues    []string
	timeout   int
}

func NewConsumer(p *redis.Pool, queues []string) *Consumer {
	return &Consumer{redisPool: p, queues: queues}
}

func (c *Consumer) Dequeue() error {
	conn := c.redisPool.Get()
	defer conn.Close()
	redisArgs := []interface{}{}
	for _, q := range c.queues {
		redisArgs = append(redisArgs, q)
	}
	redisArgs = append(redisArgs, c.timeout)

	for {
		taskB, err := redis.Strings(conn.Do("BLPOP", redisArgs...))
		if err == nil {
			logrus.Println(taskB)
		} else {
			logrus.Errorln(err)
		}
	}
}
