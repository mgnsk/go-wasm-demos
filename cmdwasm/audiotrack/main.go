//go:build js && wasm
// +build js,wasm

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/textproto"
	"sync"
	"syscall/js"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/audio"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
	"github.com/mgnsk/go-wasm-demos/public/audiotrack"
)

func main() {
	ctx := context.TODO()

	if jsutil.IsWorker {
		s := wrpc.NewServer()
		s.Run(ctx)
	} else {
		browser()
	}
}

func browser() {
	wg := &sync.WaitGroup{}
	defer jsutil.ConsoleLog("Exiting main program")
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jsutil.ConsoleLog("Starting audio")

	startAudio(ctx)

	select {}
}

func forEachChunk(reader *textproto.Reader, cb func(audio.Chunk)) {
	d := json.NewDecoder(reader.DotReader())
	for {
		var chunk audio.Chunk
		if err := d.Decode(&chunk); err != nil {
			panic(err)
		}
		cb(chunk)
	}
}

func mustWriteChunk(writer *textproto.Writer, chunk audio.Chunk) {
	dw := writer.DotWriter()
	defer dw.Close()

	e := json.NewEncoder(dw)
	if err := e.Encode(chunk); err != nil {
		panic(err)
	}
}

// TODO currently quite inefficient processing

const (
	chunkSize      = 4 * 1024
	bufferDuration = 200 * time.Millisecond
)

func startAudio(ctx context.Context) {
	// Create worker functions.

	runner := wrpc.NewWorkerRunner()

	for i := 0; i < 4; i++ {
		// TODO the worker never gets killed.
		runner.Spawn(context.TODO(), audiotrack.IndexJS)
	}

	jsutil.ConsoleLog("Workers spawned...")

	audioDecoder := func(w io.WriteCloser, _ io.Reader) {
		jsutil.ConsoleLog("1. worker")
		// Schedule a subworker to fetch and decode the chunks.
		chunkReader, chunkWriter := io.Pipe()
		wrpc.Go(chunkWriter, nil, func(w io.WriteCloser, _ io.Reader) {
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

			for chunk := range chunks {
				mustWriteChunk(writer, chunk)
				//	_ = tb
				// TODO time.Sleep takes a lot of resources.
				// Block if necessary to to stay ahead only buffer duration.
				tb.Add(chunkDuration)
			}
		})

		// Another worker to apply gain to the chunks and let it write to master out.
		wrpc.Go(w, chunkReader, func(w io.WriteCloser, r io.Reader) {
			jsutil.ConsoleLog("3. worker")
			defer w.Close()
			reader := textproto.NewReader(bufio.NewReader(r))
			writer := textproto.NewWriter(bufio.NewWriter(w))
			defer writer.W.Flush()

			forEachChunk(reader, func(chunk audio.Chunk) {
				// Apply gain FX.
				audio.Gain(&chunk, 0.5)
				mustWriteChunk(writer, chunk)
			})
		})
	}

	// We can easily passthrough cause the textprotos are buffering.
	audioPassthrough := func(w io.WriteCloser, r io.Reader) {
		jsutil.ConsoleLog("4. worker")
		defer w.Close()
		if n, err := io.Copy(w, r); err != nil {
			panic(err)
		} else if n == 0 {
			panic("0 copy")
		}
	}

	// Master tracks.
	masterReader, masterWriter := io.Pipe()
	wrpc.Go(masterWriter, nil, audioDecoder, audioPassthrough)
	reader := textproto.NewReader(bufio.NewReader(masterReader))

	audioCtx := js.Global().Get("AudioContext").New()
	player := js.Global().Get("PCMPlayer").New(audioCtx)

	forEachChunk(reader, func(chunk audio.Chunk) {
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

		player.Call("playNext", arrLeft.Float32Array(), arrRight.Float32Array())
	})
}
