package wrpc

import (
	"fmt"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/pkg/jsutil"
	"github.com/mgnsk/go-wasm-demos/pkg/wrpcnet"
)

// ListenAndServe runs the server on worker.
func ListenAndServe() error {
	port := wrpcnet.NewMessagePort(js.Global())
	defer port.Close()

	// Notify the caller to start sending calls. We have established
	// an event listener for the worker port.
	if err := port.WriteMessage(map[string]any{}, nil); err != nil {
		return fmt.Errorf("server: error sending worker init ACK: %w", err)
	}

	for {
		data, err := port.ReadMessage()
		if err != nil {
			return fmt.Errorf("server: error reading from port: %w", err)
		}
		switch {
		case !data.Get("call").IsUndefined():
			if err := call(data); err != nil {
				return err
			}
		default:
			jsutil.ConsoleLog("server: invalid message", data)
		}
	}
}

func call(data js.Value) error {
	name := data.Get("call").String()
	r := wrpcnet.NewMessagePort(data.Get("r"))
	w := wrpcnet.NewMessagePort(data.Get("w"))
	defer w.Close()

	f, ok := funcs[name]
	if !ok {
		return fmt.Errorf("server: remote func '%s' not found", name)
	}

	if err := f(w, r); err != nil {
		return w.WriteError(err)
	}

	return nil
}
