//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"io"
)

// RemoteCall is a function which must be statically declared
// so that it's pointer could be sent to another machine to run.
//
// Arguments:
// input is a reader which is piped into the worker's input.
// outputPort is call's output that must be closed when
// not being written into anymore.
// All writes to out block until a corresponding read from its other side.
type RemoteCall func(io.WriteCloser, io.Reader)

// Go provides a familiar interface for wRPC calls.
//
// Here are some rules:
// 1) f runs in a new goroutine on the first worker that receives it.
// 2) f can call Go with a new RemoteCall.
// Workers can then act like a mesh where any chain of stream is concurrently active
func Go(w io.WriteCloser, r io.Reader, f RemoteCall) {
	if w == nil {
		panic("Must have output")
	}

	var remoteReader, inputWriter, outputReader, remoteWriter Port

	if p, ok := r.(Port); ok {
		// Pass Port directly.
		remoteReader = p
	} else if r != nil {
		remoteReader, inputWriter = Pipe()
		go mustCopy(inputWriter, r)
	}

	if p, ok := w.(Port); ok {
		// Pass Port directly.
		remoteWriter = p
	} else {
		outputReader, remoteWriter = Pipe()
		go mustCopy(w, outputReader)
	}

	call := Call{
		rc:     f,
		reader: remoteReader,
		writer: remoteWriter,
	}

	go func() {
		// Schedule the call to first receiving worker.
		if err := GlobalScheduler.Call(context.TODO(), call); err != nil {
			panic(err)
		}
	}()
}

// GoPipe runs goroutines in a chain, piping each worker's output into next input.
func GoPipe(in io.Reader, out io.WriteCloser, calls ...RemoteCall) {
	prevOutReader := in
	for i, f := range calls {
		if i == len(calls)-1 {
			// The last worker writes directly into out.
			Go(out, prevOutReader, f)

		} else {
			pipeReader, pipeWriter := Pipe()
			Go(pipeWriter, prevOutReader, f)
			prevOutReader = pipeReader
		}
	}
}

func mustCopy(dst io.WriteCloser, src io.Reader) {
	defer dst.Close()
	if n, err := io.Copy(dst, src); err != nil {
		panic(err)
	} else if n == 0 {
		panic("copyAndClose: zero copy")
	}
}
