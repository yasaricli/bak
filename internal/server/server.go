// Package server provides the one-shot HTTP server that serves the diff page.
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
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

// Open serves htmlContent on the given port, opens the browser, and blocks
// until Ctrl-C or SIGTERM is received.
func Open(htmlContent string, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, htmlContent)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Fprintln(os.Stderr, "\nbye.")
		srv.Shutdown(context.Background()) //nolint:errcheck
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Fprintln(os.Stderr, "Serving", url, "— press Ctrl-C to quit")
	exec.Command("open", url).Start() //nolint:errcheck

	srv.ListenAndServe() //nolint:errcheck
}
