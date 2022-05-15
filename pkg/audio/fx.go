package audio

// Gain applies the multiplier to the passed chunk.
func Gain(chunk *Chunk, multiplier float32) {
	// TODO nil
	for i := 0; i < len(chunk.Samples); i++ {
		chunk.Samples[i] *= multiplier
	}
}
