// +build js,wasm

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/joomcode/errorx"
	"github.com/mgnsk/go-wasm-demos/public/benchmark"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
)

func init() {
	// Decode the javascript that loads and runs this binary.
	var err error
	wrpc.IndexJS, err = base64.StdEncoding.DecodeString(benchmark.IndexJS)
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

	// Create workers.
	numWorkers := 4
	var workers []*wrpc.Worker
	workerWg := &sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			workers = append(workers, wrpc.SpawnWorker(ctx))
		}()
	}
	workerWg.Wait()
	jsutil.Dump("Workers spawned...")

	inputReader, inputWriter := wrpc.Pipe()
	// TODO find the best chunk size
	go func() {
		for i := 0; i < 1000; i++ {
			if _, err := inputWriter.Write(bytes.Repeat([]byte("test"), 10000)); err != nil {
				panic(err)
			}
		}
		inputWriter.Close()
	}()

	pipeReader, pipeWriter := wrpc.Pipe()
	wrpc.Go(inputReader, pipeWriter, func(in io.Reader, out io.WriteCloser) {
		defer out.Close()
		if _, err := io.Copy(out, in); err != nil {
			panic(err)
		}
	})

	start := time.Now()
	b, err := ioutil.ReadAll(pipeReader)
	if err != nil {
		panic(err)
	}
	end := time.Since(start)
	mps := (float64(len(b)) / end.Seconds()) / 1024 / 1024

	jsutil.Dump("mps:", mps)

}
