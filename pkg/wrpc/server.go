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

	if err := port.WriteMessage(map[string]any{}, nil); err != nil {
		return fmt.Errorf("server: error sending worker init ACK: %w", err)
	}

	for {
		data, err := port.ReadMessage()
		if err != nil {
			return fmt.Errorf("server: error reading from port: %w", err)
		}

		call := data.Get("call")
		switch {
		case !call.IsUndefined():
			name := data.Get("call").String()
			f, ok := funcs[name]
			if !ok {
				fmt.Printf("server: remote func '%s' not found\n", name)
				continue
			}

			r := wrpcnet.NewMessagePort(data.Get("r"))
			w := wrpcnet.NewMessagePort(data.Get("w"))

			if err := f(w, r); err != nil {
				w.CloseWithError(err)
			} else {
				w.Close()
			}

		default:
			jsutil.ConsoleLog("server: invalid message", data)
		}
	}
}
