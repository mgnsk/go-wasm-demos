package audio

import (
	"github.com/mgnsk/go-wasm-demos/gen/idl/audio/audiov1"
)

// Gain applies the multiplier to the passed chunk.
func Gain(chunk *audiov1.Float32Chunk, multiplier float32) {
	// TODO nil
	for i := 0; i < len(chunk.Samples); i++ {
		chunk.Samples[i] *= multiplier
	}
}
