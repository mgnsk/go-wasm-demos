//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"io"
	"time"
)

// RemoteCall is a function which must be statically declared
// so that it's pointer could be sent to another worker to run,
// under the assumption that all workers run the same binary.
type RemoteCall func(io.WriteCloser, io.Reader)

// Go starts remote workers for each remote call and executes them in order by piping each
// call's output to the next input and letting the last worker write directly to w.
func Go(w io.WriteCloser, r io.Reader, calls ...RemoteCall) {
	// start copiers only when all calls are initiated
	cs := make([]Call, len(calls))
	prevReader := r
	for i, f := range calls {
		if i == len(calls)-1 {
			cs[i] = goOne(w, prevReader, f)
		} else {
			rc, wc := ConnPipe()
			cs[i] = goOne(wc, prevReader, f)
			prevReader = rc
		}
	}
	for _, c := range cs {
		c.ExecuteRemote()
	}
}

func goOne(w io.WriteCloser, r io.Reader, f RemoteCall) Call {
	if w == nil {
		panic("Must have output")
	}

	call := NewCall(w, r, f)

	worker, err := NewWorkerWithTimeout("index.js", 3*time.Second)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := worker.Call(ctx, call); err != nil {
		panic(err)
	}

	return call
}
