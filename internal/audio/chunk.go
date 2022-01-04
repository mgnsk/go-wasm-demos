package audio

// Chunk is a chunk of audio.
type Chunk struct {
	Index       uint64
	StreamStart uint64
	Samples     []float32
}
