//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"io"
)

// RemoteCall is a remote function that can be
// called on a worked by its string name.
//
// When RemoteCall returns, the underlying Writer is closed.
type RemoteCall func(io.Writer, io.Reader)

var calls = map[string]RemoteCall{}

// Handle registers a remote call with name.
func Handle(name string, call RemoteCall) {
	calls[name] = call
}

// Go starts remote workers for each remote call and executes them in order by piping each
// call's output to the next input and letting the last worker write back to caller.
// The returned WriteCloser is piped to call's Reader and the Reader is piped from call's Writer.
func Go(callNames ...string) (io.Reader, io.WriteCloser) {
	remoteReader, localWriter := Pipe()
	localReader, remoteWriter := Pipe()

	calls := make([]*Call, len(callNames))
	var prevReader io.Reader = remoteReader
	for i, name := range callNames {
		if i == len(callNames)-1 {
			calls[i] = NewCall(remoteWriter, prevReader, name)
		} else {
			rc, wc := Pipe()
			calls[i] = NewCall(wc, prevReader, name)
			prevReader = rc
		}
	}

	workers := make([]*Worker, len(calls))
	for i, call := range calls {
		worker, err := NewWorker("index.js")
		if err != nil {
			panic(err)
		}

		if err := worker.Call(call); err != nil {
			panic(err)
		}

		workers[i] = worker
	}

	for i, call := range calls {
		call.BeginRemote()
		worker := workers[i]
		go func() {
			defer worker.Close()
			if err := worker.Ping(); err != nil {
				panic(err)
			}
		}()
	}

	return localReader, localWriter
}
