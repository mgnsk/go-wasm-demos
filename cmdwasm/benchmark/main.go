//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
	"github.com/mgnsk/go-wasm-demos/public/benchmark"
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

	// Create workers.
	numWorkers := 4
	var workers []*wrpc.Worker
	runner := wrpc.NewWorkerRunner()

	for i := 0; i < numWorkers; i++ {
		w, err := runner.Spawn(ctx, benchmark.IndexJS)
		if err != nil {
			panic(err)
		}
		workers = append(workers, w)
	}

	jsutil.ConsoleLog("Workers spawned...")

	inputReader, inputWriter := io.Pipe()
	// TODO find the best chunk size
	go func() {
		for i := 0; i < 1000; i++ {
			if _, err := inputWriter.Write(bytes.Repeat([]byte("test"), 10000)); err != nil {
				panic(err)
			}
		}
		inputWriter.Close()
	}()

	pipeReader, pipeWriter := io.Pipe()
	wrpc.Go(pipeWriter, inputReader, func(w io.WriteCloser, r io.Reader) {
		defer w.Close()
		if _, err := io.Copy(w, r); err != nil {
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
