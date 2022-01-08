//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/wrpc"
)

func main() {
	if jsutil.IsWorker() {
		mux := http.NewServeMux()
		mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprintf(w, "Hello world from server!")
		})

		done := make(chan struct{})
		conns := make(chan net.Conn)
		wrpc.Handle("serve", func(w io.Writer, r io.Reader) {
			c := newPortConn(w, r)
			conns <- c
			<-c.Done()
		})

		s := &http.Server{
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		go func() {
			defer close(done)
			if err := s.Serve(newListener(conns)); err != nil {
				panic(err)
			}
		}()

		if err := wrpc.ListenAndServe(); err != nil {
			panic(err)
		}

		<-done
	} else {
		browser()
	}
}

func browser() {
	defer jsutil.ConsoleLog("Exiting main program")

	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				r, w := wrpc.Go("serve")
				return newPortConn(w, r), nil
			},
		},
	}

	resp, err := client.Get("http://localhost/hello")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

type workerAddr struct{}

func (a workerAddr) Network() string {
	return "worker"
}

func (a workerAddr) String() string {
	return ""
}

type portConn struct {
	w    io.Writer
	r    io.Reader
	once sync.Once
	done chan struct{}
}

var _ net.Conn = &portConn{}

func newPortConn(w io.Writer, r io.Reader) *portConn {
	return &portConn{
		w:    w,
		r:    r,
		done: make(chan struct{}),
	}
}

func (c *portConn) Done() <-chan struct{} {
	return c.done
}

func (c *portConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *portConn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

func (c *portConn) Close() error {
	c.once.Do(func() { close(c.done) })
	return nil
}

func (c *portConn) LocalAddr() net.Addr {
	return workerAddr{}
}

func (c *portConn) RemoteAddr() net.Addr {
	return workerAddr{}
}

func (c *portConn) SetDeadline(time.Time) error {
	return nil
}

func (c *portConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *portConn) SetWriteDeadline(time.Time) error {
	return nil
}

type workerListener struct {
	conns <-chan net.Conn
	done  chan struct{}
}

func newListener(conns <-chan net.Conn) *workerListener {
	return &workerListener{
		conns: conns,
		done:  make(chan struct{}),
	}
}

func (l *workerListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.conns:
		return c, nil
	case <-l.done:
		return nil, fmt.Errorf("listener closed")
	}
}

func (l *workerListener) Close() error {
	close(l.done)
	return nil
}

func (l *workerListener) Addr() net.Addr {
	return nil
}
