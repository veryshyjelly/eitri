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
// Callbacks must honor cancellation when intake is full or the UI is paused.
func New(addr string, deliver func(context.Context, core.Info) error, addTest func(context.Context, core.Test) error, submissions *Submissions) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleData(deliver))
	mux.HandleFunc("/tests", HandleAddTest(addTest))
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
		var data core.Info
		if !decodeJSON(w, r, &data) {
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

// HandleAddTest appends a sample to the active problem through the application's
// event loop. The callback returns only after the sample has been installed.
func HandleAddTest(addTest func(context.Context, core.Test) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Pointers distinguish missing/null fields from intentionally empty text.
		var data struct {
			Input  *string `json:"input"`
			Output *string `json:"output"`
		}
		if !decodeJSON(w, r, &data) {
			return
		}
		if data.Input == nil || data.Output == nil {
			http.Error(w, "input and output must both be strings (empty strings are allowed)", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := addTest(ctx, core.Test{Input: *data.Input, Output: *data.Output}); err != nil {
			if errors.Is(err, core.ErrNoProblem) {
				http.Error(w, err.Error(), http.StatusConflict)
			} else {
				http.Error(w, "test case intake unavailable", http.StatusServiceUnavailable)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]bool{"added": true})
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(value)
	if err == nil {
		var extra any
		if e := decoder.Decode(&extra); e != io.EOF {
			if e == nil {
				e = errors.New("multiple JSON values")
			}
			err = e
		}
	}
	if err == nil {
		return true
	}
	var oversized *http.MaxBytesError
	status := http.StatusBadRequest
	if errors.As(err, &oversized) {
		status = http.StatusRequestEntityTooLarge
	}
	http.Error(w, "invalid JSON: "+err.Error(), status)
	return false
}
