package audio

import (
	"physim/gen/idl/audio/audiov1"
	"time"
)

// MustMarshal marshals the audio buffer or panics.
func MustMarshal(chunk *audiov1.Float32Chunk) []byte {
	b, err := chunk.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}

// MustUnmarshal unmarshals the audio buffer or panics.
func MustUnmarshal(b []byte) *audiov1.Float32Chunk {
	chunk := &audiov1.Float32Chunk{}
	if err := chunk.Unmarshal(b); err != nil {
		panic(err)
	}
	return chunk
}

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
