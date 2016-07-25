// Copyright 2016. Colin Milhench. All rights reserved.

package timeline

import (
	"log"
	"time"
)

const (
	defaultPrecision = time.Minute
	defaultTtl       = time.Hour
)

type Config struct {
	DatabaseURL string
	Prefix      string
	Precision   time.Duration
	Ttl         time.Duration
}

func (c *Config) defaults() {
	if c.DatabaseURL == "" {
		log.Fatal("DatabaseURL required")
	}
	if c.Prefix == "" {
		c.Prefix = "metrics"
	}
	if c.Precision == 0 {
		c.Precision = defaultPrecision
	}
	if c.Ttl == 0 {
		c.Ttl = defaultTtl
	}
}
