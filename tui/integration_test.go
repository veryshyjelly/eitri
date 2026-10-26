package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/veryshyjelly/eitri/config"
	"github.com/veryshyjelly/eitri/core"
	"github.com/veryshyjelly/eitri/server"
	"github.com/veryshyjelly/eitri/watcher"
)

func TestIntegrationRunner(t *testing.T) {
	if !slices.Contains(os.Args, "eitri-integration-runner") {
		return
	}
	_, _ = io.Copy(os.Stdout, os.Stdin)
	os.Exit(0)
}

// Exercise transport, file generation, real watcher events, process execution,
// UI commands, copying, and one-shot submission through a running tea program.
func TestProblemLifecycle(t *testing.T) {
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	root := t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Filename:   config.Filename{Rules: []config.Rule{{Site: "example.com", Regex: `/problem/(\w+)`, Template: `{{index .Captures 1}}.txt`}}},
		Compiler:   config.Compiler{Source: "source.txt", Run: strconv.Quote(exe) + " -test.run=^TestIntegrationRunner$ -- eitri-integration-runner"},
		CreateFile: true, RunOnSave: true, Template: config.Template{Modifier: "initial"},
	}
	submissions := &server.Submissions{}
	watch := watcher.New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m := NewModel(ctx, root, cfg, Horizontal, watch, submissions)
	defer m.Close()
	copied := make(chan string, 1)
	m.writeClipboard = func(_ context.Context, text string) error { copied <- text; return nil }

	changed, selected := make(chan struct{}, 1), make(chan struct{}, 1)
	accepted := make(chan struct{}, 1)
	p := tea.NewProgram(NewSplash(m), tea.WithContext(ctx), tea.WithInput(nil), tea.WithOutput(io.Discard), tea.WithoutRenderer(), tea.WithoutSignalHandler(), tea.WithWindowSize(80, 24), tea.WithFilter(func(_ tea.Model, msg tea.Msg) tea.Msg {
		if result, ok := msg.(runResult); ok && core.GetFinalStatus(result.tests) == core.Accepted {
			accepted <- struct{}{}
		}
		return msg
	}))
	programDone := make(chan error, 1)
	go func() { _, err := p.Run(); programDone <- err }()
	defer func() {
		p.Quit()
		select {
		case <-programDone:
		case <-time.After(3 * time.Second):
			t.Error("program did not stop")
		}
	}()
	watchDone := make(chan error, 1)
	watchCtx, stopWatch := context.WithCancel(ctx)
	go func() {
		watchDone <- watch.Run(watchCtx, root, func(msg any) {
			switch msg.(type) {
			case watcher.Selected:
				if msg.(watcher.Selected).Path == "" {
					break
				}
				select {
				case selected <- struct{}{}:
				default:
				}
			case watcher.Changed:
				select {
				case changed <- struct{}{}:
				default:
				}
			}
			p.Send(msg)
		})
	}()
	defer func() {
		stopWatch()
		select {
		case err := <-watchDone:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(3 * time.Second):
			t.Error("watcher did not stop")
		}
	}()
	transport := server.New("", func(_ context.Context, info core.Info) error { p.Send(info); return nil }, submissions)
	w := httptest.NewRecorder()
	transport.Handler.ServeHTTP(w, httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"One","url":"https://example.com/problem/one","timeLimit":2000,"tests":[{"input":"42\n","output":"42"}]}`)))
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	path := filepath.Join(root, "one.txt")
	// Wait for registration of the actual target, skipping the initial deselection.
	waitUntil := time.Now().Add(3 * time.Second)
	for time.Now().Before(waitUntil) {
		select {
		case <-selected:
			if data, err := os.ReadFile(path); err == nil && string(data) == "initial" {
				goto ready
			}
		case <-time.After(10 * time.Millisecond):
		}
	}
	t.Fatal("source was not prepared/watched")
ready:
	if err := os.WriteFile(path, []byte("edited source"), 0644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-changed:
	case <-ctx.Done():
		t.Fatal("save not observed")
	}
	select {
	case <-accepted:
	case <-ctx.Done():
		t.Fatal("samples did not pass")
	}
	p.Send(tea.KeyPressMsg{Code: 'c'})
	select {
	case text := <-copied:
		if text != "edited source" {
			t.Fatalf("copied %q", text)
		}
	case <-ctx.Done():
		t.Fatal("copy not completed")
	}
	w = httptest.NewRecorder()
	transport.Handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/submit", nil))
	var result server.Submission
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Empty || result.SourceCode != "edited source" || result.FileName != "one.txt" {
		t.Fatal(fmt.Sprintf("submission: %+v", result))
	}
	w = httptest.NewRecorder()
	transport.Handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/submit", nil))
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || !result.Empty {
		t.Fatal("submission not consumed")
	}
}
