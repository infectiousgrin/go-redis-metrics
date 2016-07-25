// Copyright 2016. Colin Milhench. All rights reserved.

package timeline

import (
	"os"
	"testing"
	"time"
)

func TestHit(t *testing.T) {
	m := New(Config{
		DatabaseURL: os.Getenv("REDIS_SERVER"),
	})

	now := time.Now()

	cases := []struct {
		path  string
		start time.Time
	}{
		{"/test", now},
		{"/test", now.Add(-time.Second * 2)},    // 2 seconds ago
		{"/test", now.Add(-time.Minute * 2)},    // 2 minutes ago
		{"/test", now.Add(-time.Hour * 2)},      // 2 hours ago
		{"/test", now.Add(-time.Hour * 24 * 2)}, // 2 days ago
	}

	for _, c := range cases {
		if err := m.Store(c.path, c.start); err != nil {
			t.Error(err)
		}
	}

	if values, err := m.Fetch("/test", now.Add(-time.Minute*5), 10); err != nil {
		t.Error(err)
	} else {
		total := 0
		for _, v := range values {
			total += v
		}
		if total != 3 {
			t.Errorf("Expecting %d, Actualy got %d!", 3, total)
		}
	}
}
