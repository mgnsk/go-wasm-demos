//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"io"
)

// RemoteCall is a remote function that can be
// called on a worked by its string name.
type RemoteCall func(io.Writer, io.Reader)

var calls = map[string]RemoteCall{}

// Handle registers a remote call with name.
func Handle(name string, call RemoteCall) {
	calls[name] = call
}

// Go starts remote workers for each remote call and executes them in order by piping each
// call's output to the next input and letting the last worker write directly to w.
func Go(w io.Writer, r io.Reader, callNames ...string) {
	routines := make([]*webroutine, len(callNames))
	prevReader := r
	for i, name := range callNames {
		if i == len(callNames)-1 {
			routines[i] = newWebRoutine(w, prevReader, name)
		} else {
			rc, wc := Pipe()
			routines[i] = newWebRoutine(wc, prevReader, name)
			prevReader = rc
		}
	}
	for _, rt := range routines {
		go rt.ExecuteAndClose()
	}
}

type webroutine struct {
	worker *Worker
	call   *Call
}

func (r *webroutine) ExecuteAndClose() {
	defer r.worker.Close()
	r.call.ExecuteRemote()
}

func newWebRoutine(w io.Writer, r io.Reader, name string) *webroutine {
	if w == nil {
		panic("w must not be nil")
	}

	call := NewCall(w, r, name)

	worker, err := NewWorker("index.js")
	if err != nil {
		panic(err)
	}

	if err := worker.Call(call); err != nil {
		panic(err)
	}

	return &webroutine{
		worker: worker,
		call:   call,
	}
}
