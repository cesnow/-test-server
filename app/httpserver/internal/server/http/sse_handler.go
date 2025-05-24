package http

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"time"
)

func (s *Server) SSEHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		nt := r.URL.Query().Get("nt")
		if nt == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		clientChan := make(chan string, 10)
		s.sseMutex.Lock()
		s.sseClients[nt] = clientChan
		s.sseMutex.Unlock()

		defer func() {
			logx.Infof("client disconnected: %s\n", nt)
			s.sseMutex.Lock()
			delete(s.sseClients, nt)
			s.sseMutex.Unlock()
			close(clientChan)
		}()

		_, _ = fmt.Fprintf(w, "data: %s\n\n", "connected")
		w.(http.Flusher).Flush()

		for {
			select {
			case msg := <-clientChan:
				_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
				w.(http.Flusher).Flush()
			case <-r.Context().Done():
				return
			}
		}
	}
}

func (s *Server) SimulateEvents() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		message := time.Now().Format(time.RFC3339)

		s.sseMutex.RLock()
		for _, clientChan := range s.sseClients {
			select {
			case clientChan <- message:
				// sent successfully
			default:
				// channel is full, skip
			}
		}
		s.sseMutex.RUnlock()
	}
}
