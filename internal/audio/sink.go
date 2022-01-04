package audio

import (
	"sync/atomic"
)

// Sink is an audio sink that processes audio chunks using an array of transformers.
type Sink interface {
	Append(*Chunk)
	OutputTo(Sink)
	Drain() <-chan *Chunk
}

// Transformer is a function that modifies the chunk in place.
type Transformer func(*Chunk)

// TransformSink allows using audio transformers on passed in chunks.
type TransformSink struct {
	fx  []Transformer
	out chan *Chunk
}

// NewTransformSink constructor.
func NewTransformSink(fx ...Transformer) *TransformSink {
	return &TransformSink{
		fx:  fx,
		out: make(chan *Chunk),
	}
}

// Append to sink.
func (sink *TransformSink) Append(chunk *Chunk) {
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
func (sink *TransformSink) Drain() <-chan *Chunk {
	return sink.out
}

// EmptySink outputs all chunks it receives.
type EmptySink struct {
	out chan *Chunk
}

// NewEmptySink constructor.
func NewEmptySink() *EmptySink {
	return &EmptySink{
		out: make(chan *Chunk, 1),
	}
}

// Append to sink.
func (sink *EmptySink) Append(chunk *Chunk) {
	sink.out <- chunk
}

// OutputTo forwards all output to next sink.
func (sink *EmptySink) OutputTo(nextSink Sink) {
	for chunk := range sink.out {
		nextSink.Append(chunk)
	}
}

// Drain the sink.
func (sink *EmptySink) Drain() <-chan *Chunk {
	return sink.out
}

// OrderedSink allows appending audio chunks in any order
// and outputs them ordered.
type OrderedSink struct {
	streamStart uint64
	locker      *PriorityLocker
	out         chan *Chunk
}

// NewOrderedSink constructor.
func NewOrderedSink(streamStart uint64) *OrderedSink {
	return &OrderedSink{
		streamStart: streamStart,
		locker:      NewPriorityLocker(streamStart),
		out:         make(chan *Chunk, 1),
	}
}

// Append a chunk.
func (sink *OrderedSink) Append(chunk *Chunk) {
	if chunk.StreamStart != atomic.LoadUint64(&sink.streamStart) {
		panic("new stream must use a new sink")
	}

	// lock forces the order of chunks to be sorted.
	mu := sink.locker.NewLock(chunk.Index)
	mu.Lock()
	defer mu.Unlock()
	sink.out <- chunk
}

// OutputTo forwards all output to next sink.
func (sink *OrderedSink) OutputTo(nextSink Sink) {
	for chunk := range sink.out {
		nextSink.Append(chunk)
	}
}

// Drain the audio.
func (sink *OrderedSink) Drain() <-chan *Chunk {
	return sink.out
}
