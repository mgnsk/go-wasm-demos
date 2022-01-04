//go:build js && wasm
// +build js,wasm

package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
)

func main() {
	if jsutil.IsWorker() {
		wrpc.Handle("stringGeneratorWorker", stringGeneratorWorker)
		wrpc.Handle("upperCaseWorker", upperCaseWorker)
		wrpc.Handle("reverseWorker", reverseWorker)

		if err := wrpc.ListenAndServe(); err != nil {
			panic(err)
		}
	} else {
		browser()
	}
}

func stringGeneratorWorker(w io.Writer, r io.Reader) {
	fmt.Println("stated stringGeneratorWorker")

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

func upperCaseWorker(w io.Writer, r io.Reader) {
	fmt.Println("started upperCaseWorker")

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

func reverseWorker(w io.Writer, r io.Reader) {
	fmt.Println("started reverseWorker")

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

	// Schedule 3 workers to start streaming
	r, w := wrpc.Go("stringGeneratorWorker", "upperCaseWorker", "reverseWorker")

	// Specify the count of strings to be generated.
	enc := gob.NewEncoder(w)
	// Generate 10 strings
	count := 10
	if err := enc.Encode(&count); err != nil {
		panic(err)
	}

	if err := w.Close(); err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		jsutil.ConsoleLog("Main thread received:", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
