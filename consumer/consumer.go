package consumer

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/jiajunhuang/toq/task"
	"github.com/sirupsen/logrus"
)

var (
	sleepy           int
	sleepyQueue      chan time.Duration
	concurrencyQueue chan int
	concurrency      = flag.Int("concurrency", 10, "how many tasks can be executing at a time")
)

func init() {
	flag.Parse()
	sleepyQueue = make(chan time.Duration)
	concurrencyQueue = make(chan int, *concurrency)
	// initial concurrencyQueue so the first <*concurrency> task can start
	for i := 0; i < *concurrency; i++ {
		concurrencyQueue <- i
	}
}

type Dequeuer interface {
	Dequeue() error
}

type Worker func(task.Task) task.Result

type Consumer struct {
	redisPool *redis.Pool
	queues    []string
	timeout   int
	workers   map[string]Worker
	l         sync.RWMutex
}

func NewConsumer(p *redis.Pool, queues []string) *Consumer {
	return &Consumer{redisPool: p, queues: queues, workers: make(map[string]Worker)}
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

	for {
		// we set a sleepy queue here, because, by default, the for-loop will BLPOP all the existing
		// items in c.queues, but, consumer may be busy, and we have a concurrency limit, so, we use
		// sleepy queue to controll concurrency
		select {
		case sleepTime := <-sleepyQueue:
			sleepy++
			time.Sleep(sleepTime * time.Second)
		default:
			sleepy = 0
		}
		queueAndTask, err := redis.Strings(conn.Do("BLPOP", redisArgs...))
		if err != nil {
			logrus.Errorln(err)
			continue
		}

		queue, taskS := queueAndTask[0], queueAndTask[1]
		var t task.Task
		err = json.Unmarshal([]byte(taskS), &t)
		if err != nil {
			logrus.Errorf("failed to unmarshal task %s with error: %s", taskS, err)
			continue
		}

		// go to execute the task, but we should get token from concurrencyQueue first
		t.Road = append(t.Road, queue)
		go c.consume(t)
	}
}

func (c *Consumer) consume(t task.Task) {
	select {
	case concurrencyToken := <-concurrencyQueue:
		logrus.Infof("start to executing task %s", t.ID)
		defer func() {
			concurrencyQueue <- concurrencyToken
		}()
	default:
		logrus.Infof("too busy, I'm sleepy. sleep for %d seconds", 2*sleepy)
		sleepyQueue <- time.Duration(2 * sleepy)
	}
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
	c.SetResult(r)
	if r.WannaRetry && t.Retry && t.Tried < t.MaxRetries {
		t.Tried++
		t.Road = append(t.Road, fmt.Sprintf("retry_time_%d", t.Tried))
		logrus.Infof("start to retry task %s %d-ed time", t.ID, t.Tried)
		go c.consume(t) // or, just call it recursively?
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
