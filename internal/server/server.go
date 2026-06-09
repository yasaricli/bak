// Package server provides the HTTP server with real-time SSE-based live reload.
package server

import (
	"context"
	"crypto/md5"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type broadcaster struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func (b *broadcaster) add() chan string {
	ch := make(chan string, 1)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *broadcaster) remove(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
}

func (b *broadcaster) send(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func FreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port, nil
}

// Serve starts an HTTP server that serves the diff page with live-reload support.
// buildPage is called initially and polled every 2 seconds; when the output
// changes the connected browser tabs are notified via SSE and reload automatically.
func Serve(buildPage func() string, port int) {
	var (
		mu      sync.RWMutex
		current = buildPage()
		curHash = pageHash(current)
		bc      = &broadcaster{clients: make(map[chan string]struct{})}
	)

	go func() {
		for {
			time.Sleep(2 * time.Second)
			next := buildPage()
			h := pageHash(next)
			mu.Lock()
			if h != curHash {
				current = next
				curHash = h
				mu.Unlock()
				bc.send("update")
			} else {
				mu.Unlock()
			}
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		page := current
		mu.RUnlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page)
	})

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := bc.add()
		defer bc.remove(ch)

		fmt.Fprint(w, "data: connected\n\n")
		flusher.Flush()

		for {
			select {
			case msg := <-ch:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
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
	srv.ListenAndServe()              //nolint:errcheck
}

func pageHash(s string) string {
	h := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", h)
}
