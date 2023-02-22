package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/mgnsk/go-wasm-demos/pkg/jsutil"
	"github.com/mgnsk/go-wasm-demos/pkg/wrpc"
)

func main() {
	if jsutil.IsWorker() {
		wrpc.Register("call", func(io.Writer, io.Reader) error { return nil })
		wrpc.Register("echoBytes", func(w io.Writer, r io.Reader) error {
			if _, err := io.Copy(w, r); err != nil {
				panic(fmt.Errorf("echoBytes handler: %w", err))
			}
			return nil
		})

		if err := wrpc.ListenAndServe(); err != nil {
			panic(err)
		}
	} else {
		browser()
	}
}

// byteGenerator always reads a sequence of 1s.
type byteGenerator struct{}

func (byteGenerator) Read(b []byte) (int, error) {
	for i := 0; i < len(b); i++ {
		b[i] = 1
	}
	return len(b), nil
}

func browser() {
	defer jsutil.ConsoleLog("Exiting main program")

	jsutil.ConsoleLog("running echoBytes benchmark")

	initialSize := 1 * 1024
	maxSize := 1024 * 1024
	dur := 2 * time.Second

	for size := initialSize; size <= maxSize; size *= 2 {
		payload := bytes.Repeat([]byte{1}, size)

		r, w := wrpc.Call("echoBytes")

		go func() {
			start := time.Now()

			for time.Since(start) < dur {
				if n, err := w.Write(payload); err != nil {
					panic(err)
				} else if n != size {
					panic(io.ErrShortWrite)
				}
			}

			if err := w.Close(); err != nil {
				panic(err)
			}
		}()

		b, err := ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
		mps := (float64(len(b)) / dur.Seconds()) / 1024 / 1024

		jsutil.ConsoleLog("echoBytes %dK: MB/s:", size/1024, mps)
	}

	// jsutil.ConsoleLog("running call benchmark")
	//
	// initialConcurrency := 1
	// maxConcurrency := 32
	//
	// for concurrency := initialConcurrency; concurrency <= maxConcurrency; concurrency *= 2 {
	// 	start := time.Now()
	// 	result := make(chan float64)
	// 	for i := 0; i < concurrency; i++ {
	// 		go func() {
	// 			n := 0
	// 			for {
	// 				if d := time.Since(start); d > 2*time.Second {
	// 					ops := float64(n) / d.Seconds()
	// 					result <- ops
	// 					break
	// 				}
	// 				r, _ := wrpc.Call("call")
	// 				if _, err := r.Read(nil); err != nil && err != io.EOF {
	// 					panic(err)
	// 				}
	// 				n++
	// 			}
	// 		}()
	// 	}
	//
	// 	var ops float64
	// 	for i := 0; i < concurrency; i++ {
	// 		ops += <-result
	// 	}
	//
	// 	jsutil.ConsoleLog("call: concurency %d: ops:", concurrency, ops)
	// }

	jsutil.ConsoleLog("benchmark done")
}
