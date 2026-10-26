package watcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
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
	if err := os.WriteFile(target, []byte("saved after recovery"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := receive[Changed](t, events).Path; got != target {
		t.Fatalf("save after recovery: got %s", got)
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

type registrationHooks struct {
	add, remove func(string) error
}

func (h registrationHooks) Add(path string) error    { return h.add(path) }
func (h registrationHooks) Remove(path string) error { return h.remove(path) }

func TestWatchParentRescansAfterRegistration(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "later", "nested", "source.txt")
	var added, removed []string
	fs := registrationHooks{
		add: func(path string) error {
			added = append(added, path)
			if path == root {
				// These creates happen before Add installs the ancestor watch.
				// No filesystem notification can be relied upon to discover them.
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					return err
				}
				return os.WriteFile(target, []byte("created during registration"), 0644)
			}
			return nil
		},
		remove: func(path string) error {
			removed = append(removed, path)
			return nil
		},
	}
	directory, changed, err := watchParent(fs, target, "")
	if err != nil {
		t.Fatal(err)
	}
	if directory != filepath.Dir(target) || !changed {
		t.Fatalf("watch stuck at %s, changed %v", directory, changed)
	}
	if !slices.Equal(added, []string{root, filepath.Dir(target)}) || !slices.Equal(removed, []string{root}) {
		t.Fatalf("added %v, removed %v", added, removed)
	}
	_, changed, err = watchParent(fs, target, directory)
	if err != nil || changed || len(added) != 2 {
		t.Fatalf("stable watch was registered again: changed %v, error %v, added %v", changed, err, added)
	}
}

func TestWatchParentDisappearsDuringRegistration(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "later", "source.txt")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	fs := registrationHooks{
		add: func(path string) error {
			if path == filepath.Dir(target) {
				if err := os.RemoveAll(path); err != nil {
					return err
				}
				return os.ErrNotExist
			}
			return nil
		},
		remove: func(string) error { return nil },
	}
	directory, changed, err := watchParent(fs, target, "")
	if err != nil || !changed || directory != root {
		t.Fatalf("watch %s, changed %v, error %v", directory, changed, err)
	}
}

func TestWatchParentRemovingDeletedWatch(t *testing.T) {
	for _, removeErr := range []error{fsnotify.ErrNonExistentWatch, os.ErrNotExist, syscall.EINVAL, os.ErrPermission} {
		t.Run(removeErr.Error(), func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "later", "source.txt")
			old := filepath.Join(root, "deleted")
			fs := registrationHooks{
				add: func(string) error { return nil },
				remove: func(path string) error {
					if path != old {
						t.Fatalf("removing %s, want %s", path, old)
					}
					return removeErr
				},
			}
			directory, changed, err := watchParent(fs, target, old)
			if directory != root || !changed {
				t.Fatalf("watch %s, changed %v", directory, changed)
			}
			if errors.Is(removeErr, os.ErrPermission) {
				if !errors.Is(err, os.ErrPermission) {
					t.Fatalf("cleanup failure was hidden: %v", err)
				}
			} else if err != nil {
				t.Fatalf("already-removed watch caused an error: %v", err)
			}
		})
	}
}
