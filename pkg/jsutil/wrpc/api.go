// +build js,wasm

package wrpc

import (
	"context"
	"io"
	"time"
)

// RemoteCall is a function which must be statically declared
// so that it's pointer could be sent to another machine to run.
//
// Arguments:
// input is a reader which is piped into the worker's input.
// outputPort is call's output that must be closed when
// not being written into anymore.
// All writes to out block until a corresponding read from its other side.
type RemoteCall func(in io.Reader, out io.WriteCloser)

// Go provides a familiar interface for wRPC calls.
//
// Here are some rules:
// 1) f runs in a new goroutine on the first worker that receives it.
// 2) f can call Go with a new RemoteCall.
// Workers can then act like a mesh where any chain of stream is concurrently active
func Go(in io.Reader, out io.WriteCloser, f RemoteCall) {
	if out == nil {
		panic("Must have output")
	}

	var remoteReader, inputWriter, outputReader, remoteWriter *MessagePort

	if p, ok := in.(*MessagePort); ok {
		// Pass MessagePort directly.
		remoteReader = p
	} else if in != nil {
		remoteReader, inputWriter = Pipe()
		go func() {
			ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
			defer cancel()

			select {
			case <-inputWriter.RemoteReady():
			case <-ctx.Done():
				panic("timeout waiting for inputWriter ready")
			}

			defer inputWriter.Close()
			if n, err := io.Copy(inputWriter, in); err != nil {
				panic(err)
			} else if n == 0 {
				panic("0 write to inputWriter")
			}
		}()
	}

	if p, ok := out.(*MessagePort); ok {
		// Pass MessagePort directly.
		remoteWriter = p
	} else {
		outputReader, remoteWriter = Pipe()
		go func() {
			defer out.Close()
			if n, err := io.Copy(out, outputReader); err != nil {
				panic(err)
			} else if n == 0 {
				panic("copy to outputReader: 0 bytes")
			}
		}()
	}

	call := Call{
		RemoteCall: f,
		Input:      remoteReader,
		Output:     remoteWriter,
	}

	go func() {
		// Schedule the call to first receiving worker.
		if err := GlobalScheduler.Call(context.TODO(), call); err != nil {
			panic(err)
		}
	}()
}

// GoChain runs goroutines in a chain, piping each worker's output into next input.
func GoChain(in io.Reader, out io.WriteCloser, calls ...RemoteCall) {
	prevOutReader := in
	for i, f := range calls {
		if i == len(calls)-1 {
			// The last worker writes directly into out.
			Go(prevOutReader, out, f)

		} else {
			pipeReader, pipeWriter := Pipe()
			Go(prevOutReader, pipeWriter, f)
			prevOutReader = pipeReader
		}
	}
}
