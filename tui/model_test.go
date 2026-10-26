package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/veryshyjelly/eitri/config"
	"github.com/veryshyjelly/eitri/core"
	"github.com/veryshyjelly/eitri/server"
	"github.com/veryshyjelly/eitri/watcher"
)

func testModel(t *testing.T) Model {
	t.Helper()
	m := NewModel(context.Background(), t.TempDir(), config.Config{Filename: config.Filename{Rules: []config.Rule{{Site: "example.com", Regex: `/(\w+)$`, Template: `{{index .Captures 1}}.txt`}}}}, Horizontal, watcher.New(), &server.Submissions{})
	t.Cleanup(m.Close)
	m = m.setProblem(core.Info{Name: "One", Url: "https://example.com/one", Tests: []core.Test{{Input: "1", Output: "one"}, {Input: "2", Output: "two"}}})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	return updated.(Model)
}
func press(m Model, code rune) (Model, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyPressMsg{Code: code})
	return updated.(Model), cmd
}
func TestNavigationLayoutsAndEmptySamples(t *testing.T) {
	m := testModel(t)
	m, _ = press(m, tea.KeyTab)
	if m.index != 1 {
		t.Fatal(m.index)
	}
	m, _ = press(m, tea.KeyTab)
	if m.index != 0 {
		t.Fatal(m.index)
	}
	for key, mode := range map[rune]Mode{'1': InputAnswer, '2': InputOutput, '3': InputError, '4': AnswerOutput} {
		m, _ = press(m, key)
		if m.mode != mode {
			t.Fatal(m.mode)
		}
	}
	for _, orientation := range []Orientation{Horizontal, Vertical} {
		m.orientation = orientation
		for _, size := range []tea.WindowSizeMsg{{Width: 1, Height: 1}, {Width: 10, Height: 4}, {Width: 100, Height: 40}} {
			updated, _ := m.Update(size)
			m = updated.(Model)
			if m.leftPane.W < 1 || m.leftPane.H < 1 || m.rightPane.W < 1 || m.rightPane.H < 1 {
				t.Fatal("negative layout")
			}
			_ = m.View()
		}
	}
	updated, _ := m.Update(core.Info{Name: "Empty", Url: "https://example.com/empty"})
	m = updated.(Model)
	if m.index != 0 || m.Status != core.NotAvailable {
		t.Fatalf("index %d status %s", m.index, m.Status)
	}
	m, _ = press(m, tea.KeyTab)
	_ = m.View()
	updated, _ = m.Update(core.Info{Name: "Again", Url: "https://example.com/again", Tests: []core.Test{{Input: "a"}}})
	m = updated.(Model)
	if m.index != 0 {
		t.Fatal(m.index)
	}
	_ = m.View()
}
func TestStaleRunAndProblemResults(t *testing.T) {
	m := testModel(t)
	old, oldCmd := m.startRun()
	newer, _ := old.startRun()
	if newer.Status != core.Running {
		t.Fatal(newer.Status)
	}
	// Queued obsolete work must be skipped without touching a missing source file.
	if got := oldCmd(); got != nil {
		t.Fatalf("obsolete command ran: %v", got)
	}
	updated, _ := newer.Update(runResult{generation: old.generation, tests: []core.Testcase{{Status: core.Accepted}}})
	newer = updated.(Model)
	if newer.Status != core.Running || len(newer.Tests) != 2 {
		t.Fatal("stale result installed")
	}
	updated, _ = newer.Update(runResult{generation: newer.generation, tests: []core.Testcase{{Status: core.Accepted}}})
	newer = updated.(Model)
	if newer.Status != core.Accepted {
		t.Fatal(newer.Status)
	}
	priorGeneration := newer.generation
	updated, _ = newer.Update(core.Info{Name: "New", Url: "https://example.com/new"})
	newer = updated.(Model)
	updated, _ = newer.Update(runResult{generation: priorGeneration, tests: []core.Testcase{{Status: core.Accepted}}})
	if got := updated.(Model); got.Title != "New" || got.Status != core.NotAvailable || len(got.Tests) != 0 {
		t.Fatal("old problem result installed")
	}
}
func TestSaveInvalidatesResultsAndSubmission(t *testing.T) {
	m := testModel(t)
	m.Status = core.Running
	old := m.generation
	path := filepath.Join(m.Root, "one.txt")
	updated, _ := m.Update(watcher.Changed{Path: filepath.Join(m.Root, "other.txt")})
	if updated.(Model).generation != old {
		t.Fatal("unrelated event invalidated run")
	}
	m.submissions.Publish(server.Submission{SourceCode: "old"})
	updated, _ = m.Update(watcher.Changed{Path: path})
	m = updated.(Model)
	if m.generation == old || m.Status != core.NotAvailable || !m.fileChanged {
		t.Fatal("save did not invalidate")
	}
	updated, cmd := m.Update(copyResult{generation: old, submission: server.Submission{SourceCode: "obsolete"}})
	if cmd != nil || !updated.(Model).fileChanged {
		t.Fatal("obsolete copy installed")
	}
	m.Config.RunOnSave = true
	updated, cmd = m.Update(watcher.Changed{Path: path})
	if updated.(Model).Status != core.Running || cmd == nil {
		t.Fatal("save did not schedule run")
	}
}
func TestBackgroundMessagesInOverlays(t *testing.T) {
	m := testModel(t)
	result := runResult{generation: m.generation, tests: []core.Testcase{{Status: core.Accepted}}}
	h := Help{parent: m}
	updated, _ := h.Update(result)
	if updated.(Help).parent.(Model).Status != core.Accepted {
		t.Fatal("help dropped result")
	}
	add := NewAddTestCase(m)
	updated, _ = add.Update(result)
	if updated.(*AddTestCase).parent.Status != core.Accepted {
		t.Fatal("form dropped result")
	}
	updated, _ = add.Update(core.Info{Name: "Replacement", Url: "https://example.com/replacement"})
	if updated.(Model).Title != "Replacement" {
		t.Fatal("form dropped problem")
	}
}
func TestPreparationIsAsyncAndErrorsVisible(t *testing.T) {
	m := testModel(t)
	m.Config.CreateFile = true
	m.Template.Modifier = "{{ .Title }}"
	updated, cmd := m.Update(core.Info{Name: "Prepared", Url: "https://example.com/prepared"})
	m = updated.(Model)
	if _, err := os.Stat(filepath.Join(m.Root, "prepared.txt")); !os.IsNotExist(err) {
		t.Fatal("Update performed IO")
	}
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(Model)
	content, err := os.ReadFile(filepath.Join(m.Root, "prepared.txt"))
	if err != nil || string(content) != "Prepared" {
		t.Fatalf("%s %v", content, err)
	}
	updated, _ = m.Update(runResult{generation: m.generation, tests: m.Tests, err: errors.New("build diagnostic")})
	m = updated.(Model)
	if !strings.Contains(m.errorText, "build diagnostic") {
		t.Fatal("diagnostic lost")
	}
	_ = m.View()
}
func TestTasksSerializeCancelAndClose(t *testing.T) {
	group := newTasks(context.Background())
	defer group.close()
	var active, peak atomic.Int32
	var wg sync.WaitGroup
	entered := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		group.do(group.ctx, func(ctx context.Context) tea.Msg {
			active.Add(1)
			peak.Store(1)
			close(entered)
			<-ctx.Done()
			active.Add(-1)
			return nil
		})
	}()
	<-entered
	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			group.do(group.ctx, func(context.Context) tea.Msg {
				count := active.Add(1)
				if count > 1 {
					peak.Store(count)
				}
				active.Add(-1)
				return nil
			})
		}()
	}
	done := make(chan struct{})
	go func() { group.close(); wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("tasks leaked")
	}
	if peak.Load() != 1 {
		t.Fatal("overlapping tasks")
	}
	called := false
	group.do(context.Background(), func(context.Context) tea.Msg { called = true; return nil })
	if called {
		t.Fatal("work admitted after close")
	}
}

func TestParentCancellationReachesCommands(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	m := NewModel(parent, t.TempDir(), config.Config{}, Horizontal, watcher.New(), &server.Submissions{})
	defer m.Close()
	cancel()
	if !errors.Is(m.problemCtx.Err(), context.Canceled) {
		t.Fatal("application shutdown did not cancel operations")
	}
}

func TestMouseAndLayoutKeys(t *testing.T) {
	m := testModel(t)
	m.Tests[0].Input = strings.Repeat("line\n", 100)
	m = m.updatePanes()
	updated, _ := m.Update(tea.MouseWheelMsg{X: m.leftPane.X, Y: m.leftPane.Y, Button: tea.MouseWheelDown})
	m = updated.(Model)
	if m.leftViewPort.YOffset() == 0 {
		t.Fatal("mouse did not scroll")
	}
	updated, _ = m.Update(tea.MouseWheelMsg{X: 0, Y: 0, Button: tea.MouseWheelUp})
	m = updated.(Model)
	if m.index != 1 || !strings.Contains(m.leftViewPort.GetContent(), "2") {
		t.Fatal("mouse navigation failed")
	}
	for _, tt := range []struct {
		code rune
		want Orientation
	}{{'v', Vertical}, {'h', Horizontal}} {
		updated, _ = m.Update(tea.KeyPressMsg{Code: tt.code, Mod: tea.ModCtrl})
		m = updated.(Model)
		if m.orientation != tt.want {
			t.Fatal("layout key failed")
		}
	}
}

func TestLargeOutputDisplayIsBounded(t *testing.T) {
	m := testModel(t)
	output := strings.Repeat("line\n", 1000000)
	m.Tests[0].Output = output
	m.mode = InputOutput
	m = m.updatePanes()
	if !strings.Contains(m.rightViewPort.GetContent(), "display truncated") {
		t.Fatal("missing truncation indicator")
	}
	if m.Tests[0].Output != output {
		t.Fatal("display truncation changed judge output")
	}
	if lines := strings.Count(m.rightViewPort.GetContent(), "\n"); lines > 2001 {
		t.Fatalf("unbounded display: %d lines", lines)
	}
}
