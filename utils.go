package main

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
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
