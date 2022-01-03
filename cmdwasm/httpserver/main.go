//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
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
	ctx := context.TODO()

	if jsutil.IsWorker() {
		if err := wrpc.ListenAndServe(ctx); err != nil {
			panic(err)
		}
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
				wrpc.Go(remoteConn, remoteConn, serve)
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

func serve(w io.WriteCloser, r io.Reader) {
	defer w.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "Hello world from server!")
	})

	s := &http.Server{
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.Serve(newListener(w, r)); err != nil {
		fmt.Println(err)
	}
}

type workerListener struct {
	conn net.Conn
	once sync.Once
	done chan struct{}
}

func newListener(w io.WriteCloser, r io.Reader) *workerListener {
	c1, c2 := net.Pipe()
	go func() {
		io.Copy(c1, r)
	}()
	go func() {
		defer w.Close()
		io.Copy(w, c1)
	}()
	return &workerListener{
		conn: c2,
		done: make(chan struct{}),
	}
}

func (l *workerListener) Accept() (conn net.Conn, err error) {
	l.once.Do(func() {
		conn = l.conn
	})
	if conn != nil {
		return conn, nil
	}
	<-l.done
	return nil, fmt.Errorf("listener closed")
}

func (l *workerListener) Close() error {
	close(l.done)
	return nil
}

func (l *workerListener) Addr() net.Addr {
	return nil
}
