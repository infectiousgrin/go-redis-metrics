// Copyright 2016. Colin Milhench. All rights reserved.

package counter

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Metric struct {
	Count int64 `redis:"count"`
	Sum   int64 `redis:"sum"`
	Min   int64 `redis:"min"`
	Max   int64 `redis:"max"`
	SumSq int64 `redis:"sumsq"`
	Avg   float64
	Sdv   float64
}

type Counter struct {
	Config
	pool *redis.Pool
}

// New creates a new Counter instance.
func New(config Config) Counter {
	config.defaults()
	return Counter{
		Config: config,
		pool: &redis.Pool{
			MaxIdle:     2,
			MaxActive:   runtime.NumCPU() * 2,
			IdleTimeout: 3 * time.Minute,
			Dial: func() (redis.Conn, error) {
				c, err := redis.DialURL(config.DatabaseURL)
				if err != nil {
					return nil, err
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}
}

func (m *Counter) uuid() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (m *Counter) Incr(name string, latency int) error {
	conn := m.pool.Get()
	defer conn.Close()

	// ZSET metrics:name
	// -----------------
	// member | score
	// -----------------
	// min    | 0ms
	// max    | 0ms
	// sum    | 0ms
	// count  | 0

	// * We get 1 Hit that took 4 milliseconds...
	// ZINCRBY metrics:name 1 count
	// ZINCRBY metrics:name 4 sum
	// ZINCRBY tmp:a 4 min
	// ZINCRBY tmp:b 4 max
	// ZUNIONSTORE metrics:name 2 metrics:name tmp:a AGGREGATE MIN
	// ZUNIONSTORE metrics:name 2 metrics:name tmp:b AGGREGATE MAX
	// DEL tmp:a tmp:b

	rand.Seed(time.Now().Unix())
	var (
		key    = fmt.Sprintf("%s:%s", m.Prefix, name)
		keyMin = fmt.Sprintf("tmp:%s", m.uuid())
		keyMax = fmt.Sprintf("tmp:%s", m.uuid())
	)
	conn.Send("MULTI")
	conn.Send("ZINCRBY", key, 1, "count")
	conn.Send("ZINCRBY", key, latency, "sum")
	conn.Send("ZINCRBY", key, latency*latency, "sumsq")
	conn.Send("ZINCRBY", keyMin, latency, "min")
	conn.Send("ZINCRBY", keyMax, latency, "max")
	conn.Send("ZUNIONSTORE", key, 2, key, keyMin, "AGGREGATE", "MIN")
	conn.Send("ZUNIONSTORE", key, 2, key, keyMax, "AGGREGATE", "MAX")
	conn.Send("DEL", keyMin, keyMax)
	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}
	return nil
}

func (m *Counter) Get(name string) (Metric, error) {
	conn := m.pool.Get()
	defer conn.Close()

	var (
		key = fmt.Sprintf("%s:%s", m.Prefix, name)
		val = Metric{}
	)
	values, err := redis.Values(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		return val, err
	}
	if len(values) == 0 {
		return val, nil
	}
	if err := redis.ScanStruct(values, &val); err != nil {
		return val, err
	}
	val.Avg = float64(val.Sum) / float64(val.Count)
	n := float64(val.SumSq) - math.Pow(float64(val.Sum), 2)/float64(val.Count)
	val.Sdv = math.Sqrt(n / math.Max(float64(val.Count)-1, 1))
	return val, nil
}
