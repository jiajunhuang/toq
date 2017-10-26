package producer

import (
	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/jiajunhuang/toq/task"
	"github.com/sirupsen/logrus"
)

type Enququer interface {
	Enqueue(queue string, t task.Task) error
}

type Producer struct {
	redisPool *redis.Pool
}

func NewProducer(p *redis.Pool) *Producer {
    return &Producer{p}
}

func (p *Producer) Enqueue(queue string, t task.Task) error {
	argsB, err := json.Marshal(t)
	if err != nil {
		logrus.Errorf("Failed to marshal task: %s", t.ID)
		return err
	}
	args := string(argsB)
	conn := p.redisPool.Get()
	defer conn.Close()

	if err := conn.Send("RPUSH", queue, args); err != nil {
		return err
	}
	return nil
}
