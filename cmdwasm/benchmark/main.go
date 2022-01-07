//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
)

func main() {
	if jsutil.IsWorker() {
		wrpc.Handle("call", func(io.Writer, io.Reader) {})

		wrpc.Handle("echoBytes", func(w io.Writer, r io.Reader) {
			if _, err := io.Copy(w, r); err != nil {
				panic(err)
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
	defer jsutil.ConsoleLog("Exiting main program")

	jsutil.ConsoleLog("running echoBytes benchmark")

	r, w := wrpc.Go("echoBytes")
	go func() {
		defer w.Close()
		t := time.NewTimer(5 * time.Second)
		for {
			select {
			case <-t.C:
				return
			default:
				// write 64K bytes each time
				if _, err := w.Write(bytes.Repeat([]byte{0}, 64000)); err != nil {
					panic(err)
				}
			}
		}
	}()

	start := time.Now()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	dur := time.Since(start)
	mps := (float64(len(b)) / dur.Seconds()) / 1024 / 1024

	jsutil.ConsoleLog("echoBytes: MB/s:", mps)

	jsutil.ConsoleLog("running call benchmark")

	r, w = wrpc.Go("call")
	t := time.NewTimer(5 * time.Second)

	n := 0
	start = time.Now()
	for {
		select {
		case <-t.C:
			dur := time.Since(start)
			ops := float64(n) / dur.Seconds()
			jsutil.ConsoleLog("call: ops:", ops)
			return
		default:
			_, w = wrpc.Go("call")
			w.Close()
			n++
		}
	}
}
