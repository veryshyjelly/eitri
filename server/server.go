// Package server implements the localhost Competitive Companion transport.
// Problem processing and artifact generation belong to the application.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/veryshyjelly/eitri/core"
)

const MaxBodyBytes = 4 << 20

type Submission struct {
	Empty       bool   `json:"empty"`
	ProblemName string `json:"problemName"`
	Url         string `json:"url"`
	SourceCode  string `json:"sourceCode"`
	FileName    string `json:"fileName"`
}

// Submissions holds a single prepared artifact. Consumption is atomic: at most
// one concurrent /submit caller receives each published artifact.
type Submissions struct {
	mu      sync.Mutex
	pending *Submission
}

func (s *Submissions) Publish(value Submission) { s.mu.Lock(); defer s.mu.Unlock(); s.pending = &value }
func (s *Submissions) Clear()                   { s.mu.Lock(); defer s.mu.Unlock(); s.pending = nil }
func (s *Submissions) take() Submission {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pending == nil {
		return Submission{Empty: true}
	}
	value := *s.pending
	s.pending = nil
	return value
}

// New returns a server with a private mux and bounded request lifetimes. The
// delivery callback must honor request cancellation when its intake queue is full.
func New(addr string, deliver func(context.Context, core.Info) error, submissions *Submissions) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleData(deliver))
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(submissions.take())
	})
	return &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
}

func HandleData(deliver func(context.Context, core.Info) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		defer r.Body.Close()
		decoder := json.NewDecoder(r.Body)
		var data core.Info
		err := decoder.Decode(&data)
		if err == nil {
			var extra any
			if e := decoder.Decode(&extra); e != io.EOF {
				if e == nil {
					e = errors.New("multiple JSON values")
				}
				err = e
			}
		}
		if err != nil {
			var oversized *http.MaxBytesError
			status := http.StatusBadRequest
			if errors.As(err, &oversized) {
				status = http.StatusRequestEntityTooLarge
			}
			http.Error(w, "invalid problem JSON: "+err.Error(), status)
			return
		}
		u, err := url.Parse(data.Url)
		if strings.TrimSpace(data.Name) == "" || err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || data.TimeLimit > uint64((24*time.Hour)/time.Millisecond) {
			http.Error(w, "problem requires a name, an HTTP(S) URL, and a time limit of at most 24 hours", http.StatusBadRequest)
			return
		}
		if err := deliver(r.Context(), data); err != nil {
			http.Error(w, "problem intake unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
