package wrpc

import (
	"io"
	"sync"

	"github.com/mgnsk/go-wasm-demos/pkg/wrpcnet"
)

// Register registers a remote function.
func Register(name string, f func(io.Writer, io.Reader) error) {
	funcs[name] = f
}

// Call executes functions by chaining them and piping each function's output into the next.
//
// The returned WriteCloser is piped to the first func's Reader and
// the returned Reader is piped from the last func's Writer.
//
// The returned Reader returns the first error from any function
// or io.EOF when all functions finish.
func Call(names ...string) (io.Reader, io.WriteCloser) {
	remoteReader, localWriter := wrpcnet.Pipe()

	for _, name := range names {
		p1, p2 := wrpcnet.Pipe()

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

var funcs = map[string]func(io.Writer, io.Reader) error{}
