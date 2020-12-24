package audio

import (
	"sync/atomic"

	"github.com/mgnsk/go-wasm-demos/gen/idl/audio/audiov1"
)

// Sink is an audio sink that processes audio chunks using an array of transformers.
type Sink interface {
	Append(*audiov1.Float32Chunk)
	OutputTo(Sink)
	Drain() <-chan *audiov1.Float32Chunk
}

// Transformer is a function that modifies the chunk in place.
type Transformer func(*audiov1.Float32Chunk)

// TransformSink allows using audio transformers on passed in chunks.
type TransformSink struct {
	fx  []Transformer
	out chan *audiov1.Float32Chunk
}

// NewTransformSink constructor.
func NewTransformSink(fx ...Transformer) *TransformSink {
	return &TransformSink{
		fx:  fx,
		out: make(chan *audiov1.Float32Chunk),
	}
}

// Append to sink.
func (sink *TransformSink) Append(chunk *audiov1.Float32Chunk) {
	// apply all transforms
	for _, tr := range sink.fx {
		tr(chunk)
	}
	sink.out <- chunk
}

// OutputTo forwards all output to next sink.
func (sink *TransformSink) OutputTo(nextSink Sink) {
	for chunk := range sink.out {
		nextSink.Append(chunk)
	}
}

// Drain the sink.
func (sink *TransformSink) Drain() <-chan *audiov1.Float32Chunk {
	return sink.out
}

// EmptySink outputs all chunks it receives.
type EmptySink struct {
	out chan *audiov1.Float32Chunk
}

// NewEmptySink constructor.
func NewEmptySink() *EmptySink {
	return &EmptySink{
		out: make(chan *audiov1.Float32Chunk, 1),
	}
}

// Append to sink.
func (sink *EmptySink) Append(chunk *audiov1.Float32Chunk) {
	sink.out <- chunk
}

// OutputTo forwards all output to next sink.
func (sink *EmptySink) OutputTo(nextSink Sink) {
	for chunk := range sink.out {
		nextSink.Append(chunk)
	}
}

// Drain the sink.
func (sink *EmptySink) Drain() <-chan *audiov1.Float32Chunk {
	return sink.out
}

// OrderedSink allows appending audio chunks in any order
// and outputs them ordered.
type OrderedSink struct {
	streamStart uint64
	locker      *PriorityLocker
	out         chan *audiov1.Float32Chunk
}

// NewOrderedSink constructor.
func NewOrderedSink(streamStart uint64) *OrderedSink {
	return &OrderedSink{
		streamStart: streamStart,
		locker:      NewPriorityLocker(streamStart),
		out:         make(chan *audiov1.Float32Chunk, 1),
	}
}

// Append a chunk.
func (sink *OrderedSink) Append(chunk *audiov1.Float32Chunk) {
	if chunk.StreamStart != atomic.LoadUint64(&sink.streamStart) {
		panic("new stream must use a new sink")
	}

	// lock forces the order of chunks to be sorted.
	lock, unlock := sink.locker.GetLock(chunk.Index)
	lock()
	sink.out <- chunk
	unlock()
}

// OutputTo forwards all output to next sink.
func (sink *OrderedSink) OutputTo(nextSink Sink) {
	for chunk := range sink.out {
		nextSink.Append(chunk)
	}
}

// Drain the audio.
func (sink *OrderedSink) Drain() <-chan *audiov1.Float32Chunk {
	return sink.out
}
