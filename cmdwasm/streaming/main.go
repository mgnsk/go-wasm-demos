//go:build js && wasm
// +build js,wasm

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
)

func main() {
	ctx := context.TODO()

	if jsutil.IsWorker() {
		wrpc.Handle("stringGeneratorWorker", stringGeneratorWorker)
		wrpc.Handle("upperCaseWorker", upperCaseWorker)
		wrpc.Handle("reverseWorker", reverseWorker)
		if err := wrpc.ListenAndServe(ctx); err != nil {
			panic(err)
		}
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

	fmt.Printf("Will generate %d strings\n", n)

	bufOut := bufio.NewWriter(w)
	defer bufOut.Flush()

	for i := 0; i < int(n); i++ {
		str := "Data " + fmt.Sprintf("%f", rand.Float64()) + "\n"
		if n, err := bufOut.WriteString(str); err != nil {
			panic(err)
		} else if n == 0 {
			panic("bufOut: 0 write")
		}
		fmt.Printf("Generated %s\n", str)
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
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func browser() {
	defer jsutil.ConsoleLog("Exiting main program")

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
	outputReader, outputWriter := io.Pipe()

	// Schedule 3 workers to start streaming
	wrpc.Go(outputWriter, b, "stringGeneratorWorker", "upperCaseWorker", "reverseWorker")

	scanner := bufio.NewScanner(outputReader)
	for scanner.Scan() {
		jsutil.ConsoleLog("Main thread received:", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
