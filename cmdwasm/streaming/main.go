//go:build js && wasm
// +build js,wasm

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
	"github.com/mgnsk/go-wasm-demos/public/streaming"
)

func init() {
	// Decode the javascript that loads and runs this binary.
	var err error
	wrpc.IndexJS, err = base64.StdEncoding.DecodeString(streaming.IndexJS)
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.TODO()

	if jsutil.IsWorker {
		wrpc.RunServer(ctx)
	} else {
		browser()
	}
}

func stringGeneratorWorker(w io.WriteCloser, r io.Reader) {
	fmt.Println("stated stringGeneratorWorker")

	defer w.Close()

	// decode args
	dec := gob.NewDecoder(r)
	var n int
	if err := dec.Decode(&n); err != nil && err != io.EOF {
		panic(err)
	}

	bufOut := bufio.NewWriter(w)
	defer bufOut.Flush()

	for i := 0; i < int(n); i++ {
		str := "Test data test data " + fmt.Sprintf("%f", rand.Float64()) + "\n"
		if n, err := bufOut.WriteString(str); err != nil {
			panic(err)
		} else if n == 0 {
			panic("bufOut: 0 write")
		}
		bufOut.Flush()
		time.Sleep(500 * time.Millisecond)
	}
}

func upperCaseWorker(w io.WriteCloser, r io.Reader) {
	fmt.Println("started upperCaseWorker")

	defer w.Close()

	scanner := bufio.NewScanner(r)
	bufOut := bufio.NewWriter(w)
	defer bufOut.Flush()

	for scanner.Scan() {
		converted := strings.ToTitle(scanner.Text()) + "\n"
		if n, err := bufOut.WriteString(converted); err != nil {
			panic(err)
		} else if n == 0 {
			panic("bufOut 0 write")
		}
		bufOut.Flush()
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func reverseWorker(w io.WriteCloser, r io.Reader) {
	fmt.Println("started reverseWorker")

	defer w.Close()

	scanner := bufio.NewScanner(r)
	bufOut := bufio.NewWriter(w)
	defer bufOut.Flush()

	reverse := func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}

	for scanner.Scan() {
		reversed := reverse(scanner.Text()) + "\n"
		if n, err := bufOut.WriteString(reversed); err != nil {
			panic(err)
		} else if n == 0 {
			panic("bufOut 0 write")
		}
		bufOut.Flush()
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func browser() {
	defer jsutil.ConsoleLog("Exiting main program")

	numWorkers := 3
	var workers []*wrpc.Worker
	runner := &wrpc.WorkerRunner{}
	for i := 0; i < numWorkers; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		w, err := runner.Spawn(ctx)
		if err != nil {
			panic(err)
		}
		cancel()

		workers = append(workers, w)
	}

	jsutil.Dump("Workers spawned...")

	jsutil.ConsoleLog("Executing streaming chain call")

	// Specify the count of strings to be generated.
	b := &bytes.Buffer{}
	enc := gob.NewEncoder(b)
	// Generate 10 strings
	count := 10
	if err := enc.Encode(&count); err != nil {
		panic(err)
	}

	// Read final output from outputReader.
	// Passes outputWriter to the last worker in chain.
	outputReader, outputWriter := wrpc.Pipe()

	// Schedule 3 workers to start streaming
	wrpc.GoPipe(b, outputWriter, stringGeneratorWorker, upperCaseWorker, reverseWorker)

	time.Sleep(time.Second)

	scanner := bufio.NewScanner(outputReader)
	for scanner.Scan() {
		data := scanner.Text()
		jsutil.ConsoleLog("Main thread received:", data)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	jsutil.Dump("stream ended")

	for _, w := range workers {
		w.Terminate()
	}
	// Wait for javascript to work
	time.Sleep(3 * time.Second)
}
