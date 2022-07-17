package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/mgnsk/go-wasm-demos/pkg/jsutil"
	"github.com/mgnsk/go-wasm-demos/pkg/wrpc"
)

func main() {
	if jsutil.IsWorker() {
		wrpc.Register("stringGeneratorWorker", stringGeneratorWorker)
		wrpc.Register("upperCaseWorker", upperCaseWorker)
		wrpc.Register("reverseWorker", reverseWorker)

		if err := wrpc.ListenAndServe(); err != nil {
			panic(err)
		}
	} else {
		browser()
	}
}

func stringGeneratorWorker(w io.Writer, r io.Reader) error {
	fmt.Println("stated stringGeneratorWorker")

	// decode args
	dec := gob.NewDecoder(r)
	var n int
	if err := dec.Decode(&n); err != nil && err != io.EOF {
		return err
	}

	fmt.Printf("Will generate %d strings\n", n)

	bufOut := bufio.NewWriter(w)

	for i := 0; i < n; i++ {
		str := "Data " + fmt.Sprintf("%f", rand.Float64()) + "\n"
		if n, err := bufOut.WriteString(str); err != nil {
			return err
		} else if n == 0 {
			return fmt.Errorf("bufOut: 0 write")
		}
		fmt.Printf("Generated %s\n", str)
	}

	return bufOut.Flush()
}

func upperCaseWorker(w io.Writer, r io.Reader) error {
	fmt.Println("started upperCaseWorker")

	scanner := bufio.NewScanner(r)
	bufOut := bufio.NewWriter(w)

	for scanner.Scan() {
		converted := strings.ToTitle(scanner.Text()) + "\n"
		if n, err := bufOut.WriteString(converted); err != nil {
			return err
		} else if n == 0 {
			return fmt.Errorf("bufOut 0 write")
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return bufOut.Flush()
}

func rev(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func reverseWorker(w io.Writer, r io.Reader) error {
	fmt.Println("started reverseWorker")

	scanner := bufio.NewScanner(r)
	bufOut := bufio.NewWriter(w)

	for scanner.Scan() {
		reversed := rev(scanner.Text()) + "\n"
		if n, err := bufOut.WriteString(reversed); err != nil {
			return err
		} else if n == 0 {
			return fmt.Errorf("bufOut 0 write")
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if err := bufOut.Flush(); err != nil {
		return err
	}

	return fmt.Errorf("testing errors")
}

func browser() {
	defer jsutil.ConsoleLog("Exiting main program")

	// Schedule 3 workers to start streaming
	r, w := wrpc.Call("stringGeneratorWorker", "upperCaseWorker", "reverseWorker")

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

	err := scanner.Err()
	if err == nil {
		panic("expected error")
	}

	if err.Error() != "testing errors" {
		panic("expected error")
	}
}
