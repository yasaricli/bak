package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Broker fans out HTML update payloads to connected SSE clients.
type Broker struct {
	mu       sync.Mutex
	clients  map[chan string]struct{}
	closed   bool
}

func NewBroker() *Broker {
	return &Broker{clients: make(map[chan string]struct{})}
}

func (b *Broker) subscribe() chan string {
	ch := make(chan string, 4)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		close(ch)
		return ch
	}
	b.clients[ch] = struct{}{}
	return ch
}

func (b *Broker) unsubscribe(ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
}

// Broadcast sends html to every subscriber without blocking. If a client's
// buffer is full, the oldest message is dropped to make room.
func (b *Broker) Broadcast(html string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		select {
		case ch <- html:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- html:
			default:
			}
		}
	}
}

// Shutdown closes all subscriber channels so handler goroutines exit.
func (b *Broker) Shutdown() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for ch := range b.clients {
		delete(b.clients, ch)
		close(ch)
	}
}

// HandleSSE serves a single SSE connection. Mount at GET /events.
func (b *Broker) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := b.subscribe()
	defer b.unsubscribe(ch)

	fmt.Fprint(w, ":connected\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ":keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case msg, ok := <-ch:
			if !ok {
				return
			}
			payload, err := json.Marshal(map[string]string{"html": msg})
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
