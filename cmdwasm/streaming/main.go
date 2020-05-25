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
	"sync"
	"time"

	"github.com/joomcode/errorx"
	"github.com/mgnsk/go-wasm-demos/public/streaming"
	"github.com/mgnsk/go-wasm-demos/internal/pkg/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/pkg/jsutil/wrpc"
)

type stringGenArg int

func init() {
	// Decode the javascript that loads and runs this binary.
	var err error
	wrpc.IndexJS, err = base64.StdEncoding.DecodeString(streaming.IndexJS)
	if err != nil {
		panic(err)
	}

	gob.Register(stringGenArg(0))
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

func stringGeneratorWorker(in io.Reader, out io.WriteCloser) {
	defer out.Close()

	// decode args
	dec := gob.NewDecoder(in)
	var n stringGenArg
	if err := dec.Decode(&n); err != nil && err != io.EOF {
		panic(err)
	}

	bufOut := bufio.NewWriter(out)
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

func upperCaseWorker(in io.Reader, out io.WriteCloser) {
	defer out.Close()

	scanner := bufio.NewScanner(in)
	bufOut := bufio.NewWriter(out)
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

func reverseWorker(in io.Reader, out io.WriteCloser) {
	defer out.Close()

	scanner := bufio.NewScanner(in)
	bufOut := bufio.NewWriter(out)
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
	wg := &sync.WaitGroup{}
	defer jsutil.ConsoleLog("Exiting main program")
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	numWorkers := 3
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
	// Passes remoteWriter to the last worker in chain.
	outputReader, remoteWriter := wrpc.Pipe()

	// Schedule 3 workers to start streaming
	wrpc.GoChain(b, remoteWriter, stringGeneratorWorker, upperCaseWorker, reverseWorker)

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
