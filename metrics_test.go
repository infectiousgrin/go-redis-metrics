// Copyright 2016. Colin Milhench. All rights reserved.

package metrics

import (
	"log"
	"os"

	"github.com/garyburd/redigo/redis"
)

func init() {
	c, err := redis.DialURL(os.Getenv("REDIS_SERVER"))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	if _, err := c.Do("FLUSHDB"); err != nil {
		log.Fatal(err)
	}
}
