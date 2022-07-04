package lock

import (
	"math"
	"time"
)

// BackOff is a backoff policy for retrying an operation.
type BackOff interface {
	NextBackOff() time.Duration

	// Reset to initial state.
	Reset()
}

type ConstantBackOff struct {
	Interval time.Duration
}

func (b *ConstantBackOff) Reset()                     {}
func (b *ConstantBackOff) NextBackOff() time.Duration { return b.Interval }

func NewConstantBackOff(d time.Duration) *ConstantBackOff {
	return &ConstantBackOff{Interval: d}
}

type LinearBackOff struct {
	InitialInterval time.Duration
	Counter         int
}

func (l *LinearBackOff) Reset() { l.Counter = 1 }
func (l *LinearBackOff) NextBackOff() time.Duration {
	duration := time.Duration(l.Counter) * l.InitialInterval
	l.Counter++
	return duration
}

func NewLinearBackOff(initInterval time.Duration) *LinearBackOff {
	return &LinearBackOff{InitialInterval: initInterval, Counter: 1}
}

type ExponentialBackOff struct {
	InitialInterval time.Duration
	Multiplier      float64
	Counter         int
}

func (l *ExponentialBackOff) Reset() { l.Counter = 0 }
func (l *ExponentialBackOff) NextBackOff() time.Duration {
	duration := time.Duration(math.Pow(l.Multiplier, float64(l.Counter))) * l.InitialInterval
	l.Counter++
	return duration
}

func NewExponentialBackOff(initInterval time.Duration, multiplier float64) *ExponentialBackOff {
	return &ExponentialBackOff{
		InitialInterval: initInterval,
		Multiplier:      multiplier,
		Counter:         0}
}
