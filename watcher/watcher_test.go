package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func receive[T any](t *testing.T, events <-chan any) T {
	t.Helper()
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	for {
		select {
		case event := <-events:
			if failure, ok := event.(Error); ok {
				t.Fatal(failure.Err)
			}
			if value, ok := event.(T); ok {
				return value
			}
		case <-timer.C:
			t.Fatal("timed out waiting for watcher event")
			var zero T
			return zero
		}
	}
}
func noChange(t *testing.T, events <-chan any) {
	t.Helper()
	timer := time.NewTimer(350 * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case event := <-events:
			if changed, ok := event.(Changed); ok {
				t.Fatalf("unexpected change: %s", changed.Path)
			}
		case <-timer.C:
			return
		}
	}
}
func TestWatchSavesSwitchAndShutdown(t *testing.T) {
	root := t.TempDir()
	one := filepath.Join(root, "one.txt")
	two := filepath.Join(root, "two.txt")
	write := func(path, text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write(one, "one")
	write(two, "two")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := New()
	events := make(chan any, 64)
	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx, root, func(msg any) {
			select {
			case events <- msg:
			case <-ctx.Done():
			}
		})
	}()
	w.Select("one.txt")
	if got := receive[Selected](t, events).Path; got != one {
		t.Fatalf("selection %s", got)
	}
	for i := range 10 {
		write(one, string(rune('a'+i)))
	}
	if got := receive[Changed](t, events).Path; got != one {
		t.Fatalf("change %s", got)
	}
	noChange(t, events)
	// Editors commonly write a sibling temporary file then rename over the source.
	temp := filepath.Join(root, "save.tmp")
	write(temp, "replacement")
	if err := os.Rename(temp, one); err != nil {
		t.Fatal(err)
	}
	if got := receive[Changed](t, events).Path; got != one {
		t.Fatalf("atomic change %s", got)
	}
	write(one, "pending old save")
	w.Select("two.txt")
	receive[Selected](t, events)
	noChange(t, events)
	write(one, "ignored")
	noChange(t, events)
	write(two, "new")
	if got := receive[Changed](t, events).Path; got != two {
		t.Fatalf("switched change %s", got)
	}
	if err := os.Remove(two); err != nil {
		t.Fatal(err)
	}
	receive[Changed](t, events)
	write(two, "recreated")
	receive[Changed](t, events)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watcher did not stop")
	}
}
func TestSelectBeforeRunIsLatestWins(t *testing.T) {
	w := New()
	for range 1000 {
		w.Select("old")
	}
	w.Select("latest")
	select {
	case got := <-w.selected:
		if got != "latest" {
			t.Fatalf("got %s", got)
		}
	default:
		t.Fatal("no selection")
	}
}

func TestMissingParentRecovery(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "later", "nested", "source.txt")
	w := New()
	events := make(chan any, 64)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- w.Run(ctx, root, func(msg any) {
			select {
			case events <- msg:
			case <-ctx.Done():
			}
		})
	}()
	w.Select(target)
	receive[Selected](t, events)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := receive[Changed](t, events).Path; got != target {
		t.Fatalf("got %s", got)
	}
	// Removing an entire routed parent must also recover through ancestor events.
	if err := os.RemoveAll(filepath.Join(root, "later")); err != nil {
		t.Fatal(err)
	}
	// The first change acknowledges the removal before recreating its parent.
	receive[Changed](t, events)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("returned"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := receive[Changed](t, events).Path; got != target {
		t.Fatalf("got %s", got)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watcher did not stop")
	}
}
