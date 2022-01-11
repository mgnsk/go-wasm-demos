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

func generateChunks(w io.Writer, _ io.Reader) {
	jsutil.ConsoleLog("2. worker")
	writer := textproto.NewWriter(bufio.NewWriter(w))
	defer writer.W.Flush()

	// Currently the wav decoder requires the entire file to be downloaded before it can start producing chunks.
	// chunks := audio.GetWavChunks(wavURL, chunkSize)
	chunks := audio.GenerateChunks(5*time.Second, chunkSize)

	// Buffer up to x ms into future.
	tb := audio.NewTimeBuffer(bufferDuration)
	dur := (float64(chunkSize) * float64(time.Second)) / (2 * 44100)
	chunkDuration := time.Duration(dur)
	jsutil.ConsoleLog("Chunk duration:", chunkDuration.String())

	dw := writer.DotWriter()
	defer dw.Close()

	for chunk := range chunks {
		mustWriteChunk(dw, chunk)
		//	_ = tb
		// TODO time.Sleep takes a lot of resources.
		// Block if necessary to to stay ahead only buffer duration.
		tb.Add(chunkDuration)
	}
}

func applyGain(w io.Writer, r io.Reader) {
	jsutil.ConsoleLog("3. worker")
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
}

func audioSource(w io.Writer, _ io.Reader) {
	jsutil.ConsoleLog("1. worker")
	rr, _ := wrpc.Go("generateChunks", "applyGain")
	if _, err := io.Copy(w, rr); err != nil {
		panic(err)
	}
}

func passThrough(w io.Writer, r io.Reader) {
	jsutil.ConsoleLog("4. worker")
	if n, err := io.Copy(w, r); err != nil {
		panic(err)
	} else if n == 0 {
		panic("0 copy")
	}
}

func main() {
	if jsutil.IsWorker() {
		server := wrpc.NewServer().
			WithFunc("generateChunks", generateChunks).
			WithFunc("applyGain", applyGain).
			WithFunc("audioSource", audioSource).
			WithFunc("passThrough", passThrough)

		if err := server.ListenAndServe(); err != nil {
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
	// Master track reader.
	r, _ := wrpc.Go("audioSource", "passThrough")
	dr := textproto.NewReader(bufio.NewReader(r)).DotReader()

	audioCtx := js.Global().Get("AudioContext").New()
	player := js.Global().Get("PCMPlayer").New(audioCtx)

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

		arrLeft := array.NewFromSlice(left)
		arrRight := array.NewFromSlice(right)

		player.Call("playNext", arrLeft.JSValue(), arrRight.JSValue())
	})
}
