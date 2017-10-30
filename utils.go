package main

import (
	"flag"
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	redisPasswd = flag.String("redisPasswd", "", "redis password")
	redisURI    = flag.String("redisURI", "", "redis uri")
	redisDBNum  = flag.Int("redisDBNum", 0, "redis db num, default to 0")
	maxIdle     = flag.Int("maxIdle", 1024, "max idle in redis")
	maxActive   = flag.Int("maxActive", 100, "max active connections in redis")
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

func UUID4() string {
	u, _ := uuid.NewRandom()
	return u.String()
}
