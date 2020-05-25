// +build js,wasm

package wrpc

import (
	"context"

	"github.com/mgnsk/jsutil"
)

// Scheduler schedules calls to ports.
type Scheduler struct {
	queue chan Call
}

// NewScheduler constructor.
func NewScheduler() *Scheduler {
	return &Scheduler{
		queue: make(chan Call),
	}
}

// RunScheduler starts a scheduler to schedule calls to port.
// Runs sync on a single port.
func (s *Scheduler) RunScheduler(ctx context.Context, port *MessagePort) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case call := <-s.queue:
			messages, transferables := call.getJS()
			port.PostMessage(messages, transferables)
			<-call.Output.ack
			jsutil.Dump("Scheduler call done:", call)
		}
	}
}

// Call sends the remote call to first worker who receives it.
func (s *Scheduler) Call(ctx context.Context, call Call) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.queue <- call:
	}
	return nil
}
