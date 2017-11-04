package main

import (
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/jiajunhuang/toq/consumer"
	"github.com/jiajunhuang/toq/task"
	"github.com/sirupsen/logrus"
)

var (
	concurrency = flag.Int("concurrency", 10, "how many tasks can be executing at a time")
	debug       = flag.Bool("debug", false, "debug or not")
)

func Run(t task.Task) task.Result {
	logrus.Infof("running task %s...", t.ID)
	time.Sleep(10 * time.Second)

	return task.Result{TaskID: t.ID, WannaRetry: true, State: task.ResultStateFailed, Message: "failed"}
}

func main() {
	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debugf("running with pid: %d", os.Getpid())
	}

	redisPool := NewRedisPool()

	c := consumer.NewConsumer(redisPool, []string{"test_toq_queue"}, *concurrency)
	if err := c.RegisterWorker("test_key", Run); err != nil {
		logrus.Errorf("failed to register worker with error: %s", err)
	}
	signal.Notify(c.Sig, syscall.SIGUSR1)
	c.Dequeue()

	// graceful restart
	logrus.Warnf("graceful restarting...")
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Run()
}
