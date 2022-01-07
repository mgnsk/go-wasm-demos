//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"io"
	"sync"
)

// RemoteCall is a remote function that can be
// called on a worker by its string name.
//
// When RemoteCall returns, the underlying Writer is closed.
type RemoteCall func(io.Writer, io.Reader)

var calls = map[string]RemoteCall{}

// Handle registers a remote call with name.
func Handle(name string, call RemoteCall) {
	calls[name] = call
}

// Go starts remote workers for each remote call and executes them in order by
// piping each worker into the next.
//
// The returned WriteCloser is piped to the first call's Reader and
// the returned Reader is piped from the last call's Writer.
func Go(callNames ...string) (io.Reader, io.WriteCloser) {
	remoteReader, localWriter := Pipe()
	localReader, remoteWriter := Pipe()

	calls := make([]*Call, len(callNames))
	var prevReader *MessagePort = remoteReader
	for i, name := range callNames {
		if i == len(callNames)-1 {
			calls[i] = NewCall(remoteWriter, prevReader, name)
		} else {
			rc, wc := Pipe()
			calls[i] = NewCall(wc, prevReader, name)
			prevReader = rc
		}
	}

	for _, call := range calls {
		worker := workerPool.Get().(*Worker)
		if err := worker.Call(call); err != nil {
			panic(err)
		}
		go func() {
			defer workerPool.Put(worker)
			if err := worker.Ping(); err != nil {
				panic(err)
			}
		}()
	}

	return localReader, localWriter
}

var workerPool = sync.Pool{
	New: func() interface{} {
		worker, err := NewWorker("index.js")
		if err != nil {
			panic(err)
		}
		return worker
	},
}
