package lock

import (
	"testing"
	"time"
)

func TestLinearBackOff(t *testing.T) {
	b := NewLinearBackOff(1 * time.Second)
	for _, target := range []time.Duration{1, 2, 3, 4, 5, 6, 7, 8} {
		duration := b.NextBackOff()
		if target*time.Second != duration {
			t.Errorf("error, expect:%v, got:%v\n",
				target*time.Second, duration)
			b.NextBackOff()
		}
	}
}

func TestExponentialBackOff(t *testing.T) {
	b := NewExponentialBackOff(1*time.Second, 2)
	for _, target := range []time.Duration{1, 2, 4, 8, 16, 32} {
		duration := b.NextBackOff()
		if target*time.Second != duration {
			t.Errorf("error, expect:%v, got:%v\n",
				target*time.Second, duration)
			b.NextBackOff()
		}
	}
}
