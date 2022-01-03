//go:build js && wasm
// +build js,wasm

package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"net/textproto"
	"sync"
	"syscall/js"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/audio"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
)

func main() {
	if jsutil.IsWorker() {
		wrpc.Handle("generateChunks", func(w io.WriteCloser, _ io.Reader) {
			jsutil.ConsoleLog("2. worker")
			defer w.Close()
			writer := textproto.NewWriter(bufio.NewWriter(w))
			defer writer.W.Flush()

			// Currently the wav decoder requires the entire file to be downloaded before it can start producing chunks.
			// chunks := audio.GetWavChunks(wavURL, chunkSize)
			chunks := audio.GenerateChunks(5*time.Second, chunkSize)

			// Buffer up to x ms into future.
			tb := audio.NewTimeBuffer(bufferDuration)
			dur := (float64(chunkSize) * float64(time.Second)) / (2 * 44100)
			chunkDuration := time.Duration(dur)
			jsutil.Dump("Chunk duration:", chunkDuration)

			dw := writer.DotWriter()
			defer dw.Close()

			for chunk := range chunks {
				mustWriteChunk(dw, chunk)
				//	_ = tb
				// TODO time.Sleep takes a lot of resources.
				// Block if necessary to to stay ahead only buffer duration.
				tb.Add(chunkDuration)
			}
		})

		wrpc.Handle("applyGain", func(w io.WriteCloser, r io.Reader) {
			jsutil.ConsoleLog("3. worker")
			defer w.Close()
			reader := textproto.NewReader(bufio.NewReader(r))
			writer := textproto.NewWriter(bufio.NewWriter(w))
			defer writer.W.Flush()

			dw := writer.DotWriter()
			defer dw.Close()

			dr := reader.DotReader()

			forEachChunk(dr, func(chunk audio.Chunk) {
				// Apply gain FX.
				audio.Gain(&chunk, 0.5)
				mustWriteChunk(dw, chunk)
			})
		})

		wrpc.Handle("audioSource", func(w io.WriteCloser, _ io.Reader) {
			jsutil.ConsoleLog("1. worker")
			wrpc.Go(w, nil, "generateChunks", "applyGain")
		})

		wrpc.Handle("passthrough", func(w io.WriteCloser, r io.Reader) {
			jsutil.ConsoleLog("4. worker")
			defer w.Close()
			if n, err := io.Copy(w, r); err != nil {
				panic(err)
			} else if n == 0 {
				panic("0 copy")
			}
		})

		if err := wrpc.ListenAndServe(); err != nil {
			panic(err)
		}
	} else {
		browser()
	}
}

func browser() {
	wg := &sync.WaitGroup{}
	defer jsutil.ConsoleLog("Exiting main program")
	defer wg.Wait()

	startAudio()
}

func forEachChunk(r io.Reader, cb func(audio.Chunk)) {
	for {
		var chunk audio.Chunk
		d := gob.NewDecoder(r)
		if err := d.Decode(&chunk); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			panic(err)
		}
		cb(chunk)
	}
}

func mustWriteChunk(w io.Writer, chunk audio.Chunk) {
	e := gob.NewEncoder(w)
	if err := e.Encode(chunk); err != nil {
		panic(err)
	}
}

const (
	chunkSize      = 4 * 1024
	bufferDuration = 200 * time.Millisecond
)

func startAudio() {
	// Master tracks.
	masterReader, masterWriter := io.Pipe()
	wrpc.Go(masterWriter, nil, "audioSource", "passthrough")

	audioCtx := js.Global().Get("AudioContext").New()
	player := js.Global().Get("PCMPlayer").New(audioCtx)

	dr := textproto.NewReader(bufio.NewReader(masterReader)).DotReader()

	forEachChunk(dr, func(chunk audio.Chunk) {
		// TODO: It didn't make a difference if I sent
		// the channels together or separately.
		// Should rather try an URL object approach for some MIME type.
		left := make([]float32, len(chunk.Samples)/2)
		right := make([]float32, len(chunk.Samples)/2)

		for i, j := 0, 0; i < len(chunk.Samples)/2; i, j = i+1, j+2 {
			left[i] = chunk.Samples[j]
			right[i] = chunk.Samples[j+1]
		}

		arrLeft := array.NewArrayBufferFromSlice(left)
		arrRight := array.NewArrayBufferFromSlice(right)

		player.Call("playNext", arrLeft.Float32Array().JSValue(), arrRight.Float32Array().JSValue())
	})
}
