//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
)

func main() {
	if jsutil.IsWorker() {
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

	r, w := wrpc.Go("echoBytes")
	// TODO find the best chunk size
	go func() {
		defer w.Close()
		if _, err := w.Write(bytes.Repeat([]byte("test"), 10000)); err != nil {
			fmt.Println(err)
			return
		}
	}()

	start := time.Now()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	end := time.Since(start)
	mps := (float64(len(b)) / end.Seconds()) / 1024 / 1024

	jsutil.Dump("MB/s:", mps)
}
