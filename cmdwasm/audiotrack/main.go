// +build js,wasm

package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"io"
	"net/textproto"
	"sync"
	"syscall/js"
	"time"

	"github.com/joomcode/errorx"
	"github.com/mgnsk/go-wasm-demos/gen/idl/audio/audiov1"
	"github.com/mgnsk/go-wasm-demos/pkg/audio"
	"github.com/mgnsk/go-wasm-demos/public/audiotrack"
	"github.com/mgnsk/jsutil"
	"github.com/mgnsk/jsutil/array"
	"github.com/mgnsk/jsutil/wrpc"
)

func init() {
	// Decode the javascript that loads and runs this binary.
	var err error
	wrpc.IndexJS, err = base64.StdEncoding.DecodeString(audiotrack.IndexJS)
	if err != nil {
		panic(err)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			errorx.Panic(errorx.WithPayload(errorx.InternalError.New("panic"), r))
		}
	}()

	ctx := context.TODO()

	if jsutil.IsWorker {
		wrpc.RunServer(ctx)
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

func forEachChunk(reader *textproto.Reader, cb func(*audiov1.Float32Chunk)) {
	for {
		b, err := reader.ReadDotBytes()
		if err != nil {
			panic(err)
		}

		// Strip the ending dot and process.
		chunk := audio.MustUnmarshal(b[:len(b)-1])
		cb(chunk)
	}
}

func mustWrite(w io.Writer, p []byte) {
	if n, err := w.Write(p); err != nil {
		panic(err)
	} else if n < len(p) {
		panic(err)
	} else if n == 0 {
		panic("0 write")
	}
}

func mustWriteChunk(writer *textproto.Writer, chunk *audiov1.Float32Chunk) {
	data := audio.MustMarshal(chunk)
	dw := writer.DotWriter()
	defer dw.Close()
	mustWrite(dw, data)
}

// TODO currently quite inefficient processing

const chunkSize = 4 * 1024
const bufferDuration = 200 * time.Millisecond

func createWorkers(count int) {
	workerWg := &sync.WaitGroup{}
	for i := 0; i < count; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			// TODO the worker never gets killed.
			wrpc.SpawnWorker(context.TODO())
		}()
	}
	workerWg.Wait()
}

func startAudio(ctx context.Context) {

	// Create worker functions.

	createWorkers(4)

	audioDecoder := func(_ io.Reader, out io.WriteCloser) {
		jsutil.ConsoleLog("1. worker")
		// Schedule a subworker to fetch and decode the chunks.
		chunkReader, chunkWriter := wrpc.Pipe()
		wrpc.Go(nil, chunkWriter, func(_ io.Reader, out io.WriteCloser) {
			jsutil.ConsoleLog("2. worker")
			defer out.Close()
			writer := textproto.NewWriter(bufio.NewWriter(out))

			// Currently the wav decoder requires the entire file to be downloaded before it can start producing chunks.
			chunks := audio.GetWavChunks("https://mgnsk.github.io/go-wasm-demos/public/test2.wav", chunkSize)

			// Buffer up to x ms into future.
			tb := audio.NewTimeBuffer(bufferDuration)
			dur := (float64(chunkSize) * float64(time.Second)) / (2 * 44100)
			chunkDuration := time.Duration(dur)
			jsutil.Dump("Chunk duration:", chunkDuration)

			for chunk := range chunks {
				mustWriteChunk(writer, &chunk)
				//	_ = tb
				// TODO time.Sleep takes a lot of resources.
				// Block if necessary to to stay ahead only buffer duration.
				tb.Add(chunkDuration)
			}
		})

		// Another worker to apply gain to the chunks and let it write to master out.
		wrpc.Go(chunkReader, out, func(in io.Reader, out io.WriteCloser) {
			jsutil.ConsoleLog("3. worker")
			defer out.Close()
			reader := textproto.NewReader(bufio.NewReader(in))
			writer := textproto.NewWriter(bufio.NewWriter(out))

			forEachChunk(reader, func(chunk *audiov1.Float32Chunk) {
				// Apply gain FX.
				audio.Gain(chunk, 1.5)
				mustWriteChunk(writer, chunk)
			})
		})
	}

	// We can easily passthrough cause the textprotos are buffering.
	audioPassthrough := func(in io.Reader, out io.WriteCloser) {
		jsutil.ConsoleLog("4. worker")
		defer out.Close()
		if n, err := io.Copy(out, in); err != nil {
			panic(err)
		} else if n == 0 {
			panic("0 copy")
		}
	}

	// Master tracks.
	masterReader, masterWriter := wrpc.Pipe()
	wrpc.GoChain(nil, masterWriter, audioDecoder, audioPassthrough)
	reader := textproto.NewReader(bufio.NewReader(masterReader))

	audioCtx := js.Global().Get("AudioContext").New()
	player := js.Global().Get("PCMPlayer").New(audioCtx)

	forEachChunk(reader, func(chunk *audiov1.Float32Chunk) {
		go func() {
			// TODO: It didn't make a difference if I sent
			// the channels together or separately.
			// Should rather try an URL object approach for some MIME type.
			left := make([]float32, len(chunk.Samples)/2)
			right := make([]float32, len(chunk.Samples)/2)

			for i, j := 0, 0; i < len(chunk.Samples)/2; i, j = i+1, j+2 {
				left[i] = chunk.Samples[j]
				right[i] = chunk.Samples[j+1]
			}

			arrLeft, err := array.CreateBufferFromSlice(left)
			if err != nil {
				panic(err)
			}

			arrRight, err := array.CreateBufferFromSlice(right)
			if err != nil {
				panic(err)
			}

			player.Call("playNext", arrLeft.Float32Array(), arrRight.Float32Array())
		}()
	})
}
