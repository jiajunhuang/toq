package consumer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/jiajunhuang/toq/task"
	"github.com/sirupsen/logrus"
)

type Dequeuer interface {
	Dequeue() error
}

type Worker func(task.Task) task.Result

type Consumer struct {
	redisPool   *redis.Pool
	queues      []string
	timeout     int
	workers     map[string]Worker
	l           sync.RWMutex
	concurrency chan int
	sleepy      int
	wg          sync.WaitGroup
	Sig         chan os.Signal
}

func NewConsumer(p *redis.Pool, queues []string, maxConcurrency int) *Consumer {
	c := Consumer{
		redisPool:   p,
		queues:      queues,
		workers:     make(map[string]Worker),
		timeout:     5,
		concurrency: make(chan int, maxConcurrency),
		Sig:         make(chan os.Signal, 1),
	}
	for i := 0; i < maxConcurrency; i++ {
		c.concurrency <- i
	}

	return &c
}

func (c *Consumer) RegisterWorker(key string, w Worker) error {
	c.l.Lock()
	defer c.l.Unlock()

	if _, ok := c.workers[key]; ok {
		return errors.New("the key has been registed")
	}
	c.workers[key] = w

	return nil
}

func (c *Consumer) Dequeue() error {
	conn := c.redisPool.Get()
	defer conn.Close()

	redisArgs := []interface{}{}
	for _, q := range c.queues {
		redisArgs = append(redisArgs, q)
	}
	redisArgs = append(redisArgs, c.timeout)

loop:
	for {
		// we set a sleepy queue here, because, by default, the for-loop will BLPOP all the existing
		// items in c.queues, but, consumer may be busy, and we have a concurrency limit, so, we use
		// sleepy queue to controll concurrency
		var token int
		select {
		case s := <-c.Sig:
			logrus.Warnf("received a signal %s", s)
			goto restart
		default:

		}
		select {
		case token = <-c.concurrency:
			logrus.Debugf("got token %d", token)
			c.sleepy = 0
		default:
			c.sleepy++
			logrus.Warnf("sleep %d seconds", c.sleepy*2)
			time.Sleep(time.Duration(c.sleepy*2) * time.Second)
			goto loop
		}
		queueAndTask, err := redis.Strings(conn.Do("BLPOP", redisArgs...))
		if err != nil {
			logrus.Errorln(err)
			c.concurrency <- token
			continue
		}

		queue, taskS := queueAndTask[0], queueAndTask[1]
		var t task.Task
		err = json.Unmarshal([]byte(taskS), &t)
		if err != nil {
			logrus.Errorf("failed to unmarshal task %s with error: %s", taskS, err)
			c.concurrency <- token
			continue
		}

		// go to execute the task, but we should get token from concurrencyQueue first
		t.Road = append(t.Road, queue)
		logrus.Infof("start to execute task %s with token %d", t.ID, token)
		go c.consume(t, token)
	}

restart:
	logrus.Warnf("wait for all tokens and then exit")
	c.wg.Wait()
	return nil
}

func (c *Consumer) consume(t task.Task, token int) {
	c.wg.Add(1)
	defer c.wg.Done()
	// find the worker
	w, ok := c.workers[t.Key]
	if !ok {
		logrus.Errorf("cannot find correspoding worker of task %s with key %s", t.ID, t.Key)
		r := task.Result{TaskID: t.ID, State: task.ResultStateFailed, Message: "worker not found"}
		c.SetResult(r)
		return
	}

	// run
	r := w(t)
	logrus.Infof("task %s returned a result with state %d, road: %v, wanna retry: %t, max retries: %d, tried: %d", t.ID, r.State, t.Road, r.WannaRetry, t.MaxRetries, t.Tried)
	if r.WannaRetry && t.Retry && t.Tried < t.MaxRetries {
		t.Tried++
		t.Road = append(t.Road, fmt.Sprintf("retry_time_%d", t.Tried))
		logrus.Infof("start to retry task %s %d-ed time", t.ID, t.Tried)
		go c.consume(t, token) // or, just call it recursively?
	} else {
		c.SetResult(r)
		c.concurrency <- token
	}
}

func (c *Consumer) SetResult(r task.Result) {
	conn := c.redisPool.Get()
	defer conn.Close()

	resultB, err := json.Marshal(r)
	if err != nil {
		logrus.Errorf("failed to marshal result %v", r)
	} else {
		if err := conn.Send("SETEX", fmt.Sprintf("toq:rst:%s", r.TaskID), 500, string(resultB)); err != nil {
			logrus.Errorf("failed to set result to redis: %s", err)
		} else {
			logrus.Infof("result of task %s had sent to redis and will expire in 500s", r.TaskID)
		}
	}
}
