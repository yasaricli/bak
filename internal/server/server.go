// Package server provides the long-lived HTTP server that serves the diff
// page and pushes live updates over SSE.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync/atomic"
	"syscall"
)

func FreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port, nil
}

// Open serves the HTML currently held in pageRef on /, the SSE stream on
// /events, opens the browser, and blocks until Ctrl-C or SIGTERM is received.
// pageRef must hold a string.
func Open(pageRef *atomic.Value, port int, broker *Broker) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html, _ := pageRef.Load().(string)
		fmt.Fprint(w, html)
	})
	if broker != nil {
		mux.HandleFunc("/events", broker.HandleSSE)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Fprintln(os.Stderr, "\nbye.")
		if broker != nil {
			broker.Shutdown()
		}
		srv.Shutdown(context.Background()) //nolint:errcheck
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Fprintln(os.Stderr, "Serving", url, "— press Ctrl-C to quit")
	exec.Command("open", url).Start() //nolint:errcheck

	srv.ListenAndServe() //nolint:errcheck
}
