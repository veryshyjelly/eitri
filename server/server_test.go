package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/veryshyjelly/eitri/core"
)

const validPayload = `{"name":"A","url":"https://example.com/problem/a","timeLimit":1000,"tests":[{"input":"1","output":"2"}],"batch":{"id":"x","size":1}}`

func TestAddTestValidation(t *testing.T) {
	for _, tt := range []struct {
		name, method, body string
		status             int
	}{
		{"valid", "POST", `{"input":"1\n","output":"2\n"}`, http.StatusCreated},
		{"empty strings", "POST", `{"input":"","output":""}`, http.StatusCreated},
		{"missing input", "POST", `{"output":"2"}`, http.StatusBadRequest},
		{"missing output", "POST", `{"input":"1"}`, http.StatusBadRequest},
		{"null input", "POST", `{"input":null,"output":""}`, http.StatusBadRequest},
		{"null output", "POST", `{"input":"","output":null}`, http.StatusBadRequest},
		{"wrong type", "POST", `{"input":1,"output":""}`, http.StatusBadRequest},
		{"invalid JSON", "POST", `{`, http.StatusBadRequest},
		{"trailing JSON", "POST", `{"input":"","output":""} {}`, http.StatusBadRequest},
		{"oversized", "POST", `{"input":"` + strings.Repeat("x", MaxBodyBytes) + `","output":""}`, http.StatusRequestEntityTooLarge},
		{"wrong method", "GET", ``, http.StatusMethodNotAllowed},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			srv := New("", nil, func(ctx context.Context, test core.Test) error {
				called = true
				if _, ok := ctx.Deadline(); !ok {
					t.Error("missing request deadline")
				}
				if tt.name == "valid" && (test.Input != "1\n" || test.Output != "2\n") {
					t.Errorf("test text changed: %+v", test)
				}
				return nil
			}, &Submissions{})
			w := httptest.NewRecorder()
			srv.Handler.ServeHTTP(w, httptest.NewRequest(tt.method, "/tests", strings.NewReader(tt.body)))
			if w.Code != tt.status || called != (tt.status == http.StatusCreated) {
				t.Fatalf("status %d, called %v: %s", w.Code, called, w.Body.String())
			}
			if w.Code == http.StatusCreated && (w.Header().Get("Content-Type") != "application/json" || strings.TrimSpace(w.Body.String()) != `{"added":true}`) {
				t.Fatalf("unexpected response: %v %s", w.Header(), w.Body.String())
			}
			if w.Code == http.StatusMethodNotAllowed && w.Header().Get("Allow") != "POST" {
				t.Fatal("missing Allow header")
			}
		})
	}
	for _, tt := range []struct {
		err    error
		status int
	}{{core.ErrNoProblem, http.StatusConflict}, {context.Canceled, http.StatusServiceUnavailable}, {errors.New("stopped"), http.StatusServiceUnavailable}} {
		w := httptest.NewRecorder()
		HandleAddTest(func(context.Context, core.Test) error { return tt.err })(w, httptest.NewRequest("POST", "/tests", strings.NewReader(`{"input":"","output":""}`)))
		if w.Code != tt.status {
			t.Fatalf("error %v: status %d", tt.err, w.Code)
		}
	}
}

func TestIntakeValidation(t *testing.T) {
	for _, tt := range []struct {
		name, method, body string
		status             int
	}{
		{"valid", "POST", validPayload, 200}, {"empty samples", "POST", `{"name":"A","url":"https://example.com/a"}`, 200},
		{"invalid JSON", "POST", "{", 400}, {"trailing JSON", "POST", validPayload + ` {}`, 400},
		{"missing name", "POST", `{"url":"https://example.com"}`, 400}, {"bad URL", "POST", `{"name":"A","url":"file:///tmp/source"}`, 400},
		{"oversized", "POST", `{"name":"` + strings.Repeat("x", MaxBodyBytes) + `"}`, 413},
		{"method", "GET", validPayload, 405},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			h := HandleData(func(_ context.Context, info core.Info) error {
				called = true
				if info.Name != "A" {
					t.Fatal(info.Name)
				}
				return nil
			})
			w := httptest.NewRecorder()
			h(w, httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body)))
			if w.Code != tt.status || called != (tt.status == 200) {
				t.Fatalf("status %d, called %v", w.Code, called)
			}
		})
	}
	w := httptest.NewRecorder()
	HandleData(func(context.Context, core.Info) error { return errors.New("stopped") })(w, httptest.NewRequest("POST", "/", strings.NewReader(validPayload)))
	if w.Code != 503 {
		t.Fatalf("got %d", w.Code)
	}
}
func TestConcurrentIntakeAndSubmission(t *testing.T) {
	var accepted atomic.Int64
	submissions := &Submissions{}
	srv := New("", func(context.Context, core.Info) error { accepted.Add(1); return nil }, nil, submissions)
	submissions.Publish(Submission{SourceCode: "prepared", ProblemName: "A"})
	var consumed atomic.Int64
	var wg sync.WaitGroup
	for range 40 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			srv.Handler.ServeHTTP(w, httptest.NewRequest("POST", "/", strings.NewReader(validPayload)))
			if w.Code != 200 {
				t.Errorf("intake: %d", w.Code)
			}
			w = httptest.NewRecorder()
			srv.Handler.ServeHTTP(w, httptest.NewRequest("GET", "/submit", nil))
			var result Submission
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Error(err)
				return
			}
			if !result.Empty {
				consumed.Add(1)
				if result.SourceCode != "prepared" {
					t.Error(result)
				}
			}
		}()
	}
	wg.Wait()
	if accepted.Load() != 40 || consumed.Load() != 1 {
		t.Fatalf("accepted %d, consumed %d", accepted.Load(), consumed.Load())
	}
	submissions.Publish(Submission{SourceCode: ""})
	if got := submissions.take(); got.Empty {
		t.Fatal("empty artifact treated as unavailable")
	}
	submissions.Publish(Submission{})
	submissions.Clear()
	if !submissions.take().Empty {
		t.Fatal("clear failed")
	}
}
func TestGracefulShutdown(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	srv := New(listener.Addr().String(), func(context.Context, core.Info) error { close(entered); <-release; return nil }, nil, &Submissions{})
	serveDone := make(chan error, 1)
	go func() { serveDone <- srv.Serve(listener) }()
	requestDone := make(chan error, 1)
	go func() {
		response, err := http.Post("http://"+listener.Addr().String()+"/", "application/json", strings.NewReader(validPayload))
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
		requestDone <- err
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("request never entered")
	}
	shutdownDone := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() { shutdownDone <- srv.Shutdown(ctx) }()
	close(release)
	if err := <-requestDone; err != nil {
		t.Fatal(err)
	}
	if err := <-shutdownDone; err != nil {
		t.Fatal(err)
	}
	if err := <-serveDone; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("serve: %v", err)
	}
}
