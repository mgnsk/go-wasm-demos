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

	for _, name := range callNames {
		p1, p2 := Pipe()
		call := NewCall(p1, remoteReader, name)
		remoteReader = p2

		worker := pool.Get().(*Worker)
		if err := worker.Call(call); err != nil {
			panic(err)
		}

		go func() {
			defer pool.Put(worker)
			if err := worker.Ping(); err != nil {
				panic(err)
			}
		}()
	}

	return remoteReader, localWriter
}

var pool = sync.Pool{
	New: func() interface{} {
		worker, err := NewWorker("index.js")
		if err != nil {
			panic(err)
		}
		return worker
	},
}
