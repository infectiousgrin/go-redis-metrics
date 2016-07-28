// Copyright 2016. Colin Milhench. All rights reserved.

package timeline

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Timeline struct {
	Config
	pool *redis.Pool
}

func New(config Config) *Timeline {
	config.defaults()
	return &Timeline{
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

func (m *Timeline) round(date time.Time, precision time.Duration) time.Time {
	d := float64(date.UnixNano())
	p := float64(precision)
	return time.Unix(0, int64(math.Floor(d/p)*p))
}

func (m *Timeline) Store(name string, when time.Time) error {
	conn := m.pool.Get()
	defer conn.Close()

	// HASH (hits by 1 minute, for 24 hours)
	// metrics:24h0m0s:1464825600000:/test (previous)
	// metrics:24h0m0s:1464912000000:/test (current)
	// 1464912000000 | 1
	// 1464912060000 | 3
	// ............. | 2

	// * We get 1 Hit
	// HINCRBY  metrics:<precision>:<bucket>:<name> <tick> 1
	// EXPITEAT metrics:<precision>:<bucket>:<name> <bucket+(ttl*2)>

	var (
		tick   = m.round(when, m.Precision)
		bucket = m.round(tick, m.Ttl)
		expire = bucket.Add(2 * m.Ttl)
		key    = fmt.Sprintf("%s:%s:%d:%s", m.Prefix, m.Ttl, bucket.UnixNano()/1e9, name)
	)

	conn.Send("MULTI")
	conn.Send("HINCRBY", key, tick, 1)
	conn.Send("EXPIREAT", key, expire.UnixNano()/1e9)
	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}

	return nil
}

func (m *Timeline) Fetch(name string, when time.Time, count int) (map[time.Time]int, error) {
	conn := m.pool.Get()
	defer conn.Close()

	var start = m.round(when, m.Precision)

	conn.Send("MULTI")
	for i := 0; i < count; i++ {
		var (
			tick   = start.Add(m.Precision * time.Duration(i))
			bucket = m.round(tick, m.Ttl)
			key    = fmt.Sprintf("%s:%s:%d:%s", m.Prefix, m.Ttl, bucket.UnixNano()/1e9, name)
		)
		conn.Send("HGET", key, tick)
	}

	values, err := redis.Ints(conn.Do("EXEC"))

	if err != nil {
		return nil, err
	}

	x := map[time.Time]int{}
	for i, v := range values {
		x[start.Add(m.Precision*time.Duration(i))] = v
	}

	return x, nil

}
