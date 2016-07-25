// Copyright 2016. Colin Milhench. All rights reserved.

package counter

import "log"

type Config struct {
	DatabaseURL string
	Prefix      string
}

func (c *Config) defaults() {
	if c.DatabaseURL == "" {
		log.Fatal("DatabaseURL required")
	}
	if c.Prefix == "" {
		c.Prefix = "metrics"
	}
}
