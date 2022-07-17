package wrpc

import (
	"io"
	"sync"

	"github.com/mgnsk/go-wasm-demos/pkg/jsutil"
)

// WorkerFunc is a streaming remote function.
type WorkerFunc func(io.Writer, io.Reader) error

// Register a remote function on a worker.
func Register(name string, f WorkerFunc) {
	if !jsutil.IsWorker() {
		panic("Register must be called on a worker")
	}
	funcs[name] = f
}

// Call acquires workers and executes streaming functions in order by
// piping each worker into the next.
//
// The returned WriteCloser is piped to the first func's Reader and
// the returned Reader is piped from the last func's Writer.
//
// The returned Reader returns the first error from any WorkerFunc
// or io.EOF when all functions finish.
func Call(names ...string) (io.Reader, io.WriteCloser) {
	remoteReader, localWriter := Pipe()

	for _, name := range names {
		p1, p2 := Pipe()

		w := p1
		r := remoteReader
		name := name

		worker := pool.Get().(*Worker)
		go func() {
			if err := worker.Call(w, r, name); err != nil {
				panic(err)
			}
			pool.Put(worker)
		}()

		remoteReader = p2
	}

	return remoteReader, localWriter
}

var pool = sync.Pool{
	New: func() any {
		worker, err := NewWorker("index.js")
		if err != nil {
			panic(err)
		}
		return worker
	},
}

var funcs = map[string]WorkerFunc{}
