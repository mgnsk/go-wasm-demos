//go:build js && wasm
// +build js,wasm

// Package jsutil provides general functionality for any application running on wasm.
package jsutil

import (
	"syscall/js"
)

// IsWorker returns whether we are running in a webworker.
func IsWorker() bool {
	return js.Global().Get("WorkerGlobalScope").Type() != js.TypeUndefined
}

// CreateURLObject creates an url object from javascript source.
func CreateURLObject(data interface{}, mime string) js.Value {
	blob := js.Global().Get("Blob").New([]interface{}{data}, map[string]interface{}{"type": mime})
	return js.Global().Get("URL").Call("createObjectURL", blob)
}

// ConsoleLog console.log
func ConsoleLog(args ...interface{}) {
	js.Global().Get("console").Call("log", args...)
}
