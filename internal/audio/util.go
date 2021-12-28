package audio

import (
	"time"
)

// TimeBuffer blocks a speedy caller to be ahead of realtime by max amount.
type TimeBuffer struct {
	max     time.Duration
	start   time.Time
	elapsed time.Duration
}

// NewTimeBuffer constructor.
func NewTimeBuffer(max time.Duration) *TimeBuffer {
	return &TimeBuffer{
		max:   max,
		start: time.Now(),
	}
}

// Add an amount of time to buffer the caller thinks has passed.
// It will block if necessary to slow down the caller to be
// ahead of real time by max.
func (tb *TimeBuffer) Add(elapsed time.Duration) {
	tb.elapsed += elapsed
	if tb.elapsed > tb.max {
		allowedAt := time.Now().Add(tb.max)
		bufferedAt := tb.start.Add(tb.elapsed)
		wait := bufferedAt.Sub(allowedAt)
		if wait > 0 {
			time.Sleep(wait)
			tb.start = time.Now()
			tb.elapsed = 0
		}
	}
}
