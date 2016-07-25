// Copyright 2016. Colin Milhench. All rights reserved.

package counter

import (
	"math"
	"os"
	"testing"
)

func TestIncr(t *testing.T) {
	m := New(Config{
		DatabaseURL: os.Getenv("REDIS_SERVER"),
	})

	cases := []struct {
		name    string
		latency int
	}{
		{"http://www.mathsisfun.com/data/standard-deviation-calculator.html", -5},
		{"http://www.mathsisfun.com/data/standard-deviation-calculator.html", 1},
		{"http://www.mathsisfun.com/data/standard-deviation-calculator.html", 8},
		{"http://www.mathsisfun.com/data/standard-deviation-calculator.html", 7},
		{"http://www.mathsisfun.com/data/standard-deviation-calculator.html", 2},
	}

	for _, c := range cases {
		if err := m.Incr(c.name, c.latency); err != nil {
			t.Error(err)
		}
	}
	if val, err := m.Get("http://www.mathsisfun.com/data/standard-deviation-calculator.html"); err != nil {
		t.Error(err)
	} else {
		if val.Avg != 2.6 {
			t.Errorf("Expecting %f, Actualy Got %f", 2.6, val.Avg)
		}
		var EPSILON float64 = 0.00000001
		var expected = 5.224940191
		if diff := math.Abs(val.Sdv - expected); diff > EPSILON {
			t.Errorf("Expecting %f, Actualy Got %f!", 5.224940191, val.Sdv)
		}
	}
}
