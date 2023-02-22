// Package jsutil provides general functionality for any application running on wasm.
package jsutil

import (
	"syscall/js"
)

// IsWorker returns whether the program is running in a webworker.
func IsWorker() bool {
	return js.Global().Get("WorkerGlobalScope").Type() != js.TypeUndefined
}

// CreateURLObject creates an url object.
func CreateURLObject(data any, mime string) js.Value {
	blob := js.Global().Get("Blob").New([]any{data}, map[string]any{"type": mime})
	return js.Global().Get("URL").Call("createObjectURL", blob)
}

var console = js.Global().Get("console")

// ConsoleLog console.log
func ConsoleLog(args ...any) {
	// js.Global().Get("console").Call("log", args...)
	console.Call("log", args...)
}
