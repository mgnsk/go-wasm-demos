// +build js,wasm

package wrpc

// GlobalScheduler is main scheduler to schedule to workers.
var GlobalScheduler = NewScheduler()

// CallCount specifies how many calls are currently processing.
var CallCount uint64 = 0
