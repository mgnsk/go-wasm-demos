//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"io"
	"time"
)

var calls = map[string]func(io.WriteCloser, io.Reader){}

// Handle registers a remote call with name.
func Handle(name string, call func(io.WriteCloser, io.Reader)) {
	calls[name] = call
}

// Go starts remote workers for each remote call and executes them in order by piping each
// call's output to the next input and letting the last worker write directly to w.
func Go(w io.WriteCloser, r io.Reader, callNames ...string) {
	// start copiers only when all calls are initiated
	cs := make([]Call, len(callNames))
	prevReader := r
	for i, name := range callNames {
		if i == len(callNames)-1 {
			cs[i] = goOne(w, prevReader, name)
		} else {
			rc, wc := connPipe()
			cs[i] = goOne(wc, prevReader, name)
			prevReader = rc
		}
	}
	for _, c := range cs {
		c.ExecuteRemote()
	}
}

func goOne(w io.WriteCloser, r io.Reader, name string) Call {
	if w == nil {
		panic("Must have output")
	}

	call := NewCall(w, r, name)

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
