//go:build js && wasm
// +build js,wasm

package audio

import (
	"bytes"
	"io/ioutil"
	"net/http"

	goaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// GetWavChunks fetches a wav file and returns a channel to PCM chunks.
func GetWavChunks(wavURL string, chunkSamples int) chan Chunk {
	resp, err := http.Get(wavURL)
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	decoder := wav.NewDecoder(bytes.NewReader(b))
	if !decoder.IsValidFile() {
		panic("invalid wav file")
	}

	chunks := make(chan Chunk)
	go func() {
		defer close(chunks)
		index, streamStart := uint64(0), uint64(0)
		for {
			// decode audio to pcm data
			buffer := &goaudio.IntBuffer{
				Data: make([]int, chunkSamples/4),
			}
			if n, err := decoder.PCMBuffer(buffer); err != nil {
				panic(err)
			} else if n == 0 {
				return
			}

			// copy the buffer to []float32
			f32Buffer := buffer.AsFloat32Buffer()

			chunks <- Chunk{
				Index:       index,
				StreamStart: streamStart,
				Samples:     f32Buffer.Data,
			}
			index++
		}
	}()

	return chunks
}
