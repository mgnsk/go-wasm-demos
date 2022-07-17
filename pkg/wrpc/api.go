package wrpc

import (
	"io"
	"sync"
)

// WorkerFunc is a function that can be
// executed on a worker by its string name.
type WorkerFunc func(io.Writer, io.Reader)

// Go acquires workers and executes WorkerFuncs in order by
// piping each worker into the next.
//
// The returned WriteCloser is piped to the first func's Reader and
// the returned Reader is piped from the last func's Writer.
//
// When all funcs are finished, the returned io.Reader returns io.EOF.
func Go(funcs ...string) (io.Reader, io.WriteCloser) {
	remoteReader, localWriter := Pipe()

	for _, name := range funcs {
		p1, p2 := Pipe()

		w := p1
		r := remoteReader
		name := name

		worker := pool.Get().(*Worker)
		go func() {
			if err := worker.Execute(w, r, name); err != nil {
				panic(err)
			}
			pool.Put(worker)
		}()

		remoteReader = p2
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
