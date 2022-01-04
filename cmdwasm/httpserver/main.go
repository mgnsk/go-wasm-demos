//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
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
			conns <- combine(w, r)
			<-done
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
				localConn, remoteConn := net.Pipe()
				wrpc.Go(remoteConn, remoteConn, "serve")
				return localConn, nil
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

type workerListener struct {
	conns <-chan net.Conn
	done  chan struct{}
}

func combine(w io.Writer, r io.Reader) net.Conn {
	c1, c2 := net.Pipe()
	go io.Copy(c1, r)
	go io.Copy(w, c1)
	return c2
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
