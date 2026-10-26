// Package watcher turns directory notifications into debounced source changes.
package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/veryshyjelly/eitri/config"
)

// Changed identifies the full source path so the UI can reject a queued event
// from a problem that has since been replaced.
type Changed struct{ Path string }
type Error struct{ Err error }

// Selected acknowledges that the source selection has been applied.
type Selected struct{ Path string }

// Watcher owns its fsnotify handle and selected file. Select updates are latest
// wins and nonblocking, allowing the UI to switch problems without waiting for IO.
type Watcher struct{ selected chan string }

func New() *Watcher { return &Watcher{selected: make(chan string, 1)} }
func (w *Watcher) Select(path string) {
	select {
	case w.selected <- path:
		return
	default:
	}
	select {
	case <-w.selected:
	default:
	}
	select {
	case w.selected <- path:
	default:
	}
}

// Run watches parent directories to survive atomic rename/replace saves. Timer
// state belongs entirely to this loop; switching sources discards pending saves.
func (w *Watcher) Run(ctx context.Context, root string, emit func(any)) error {
	fs, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer fs.Close()
	var path, directory string
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()
	var ticks <-chan time.Time
	stopTimer := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		ticks = nil
	}
	// When a routed parent does not exist yet, watch its nearest existing
	// ancestor and move the registration down as directories are created.
	refresh := func() (bool, error) {
		target := ""
		if path != "" {
			target = filepath.Dir(path)
			for {
				stat, err := os.Stat(target)
				if err == nil && !stat.IsDir() {
					return false, fmt.Errorf("watch parent is not a directory: %s", target)
				}
				if err == nil {
					if target == directory {
						return false, nil
					}
					err = fs.Add(target)
					if err == nil {
						break
					}
				}
				if !os.IsNotExist(err) {
					return false, err
				}
				// A directory may vanish between Stat and Add during a tree removal.
				// Move upward immediately instead of treating that expected race as fatal.
				parent := filepath.Dir(target)
				if parent == target {
					return false, err
				}
				target = parent
			}
		}
		if target == directory {
			return false, nil
		}

		old := directory
		directory = target
		if old != "" {
			if err := fs.Remove(old); err != nil && !errors.Is(err, fsnotify.ErrNonExistentWatch) && !os.IsNotExist(err) {
				return true, err
			}
		}
		return true, nil
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case selected := <-w.selected:
			stopTimer()
			next := ""
			if selected != "" {
				next, err = filepath.Abs(config.ResolvePath(root, selected))
				if err != nil {
					emit(Error{err})
					continue
				}
			}
			path = next
			if _, err := refresh(); err != nil {
				emit(Error{fmt.Errorf("watch source directory: %w", err)})
			}

			emit(Selected{Path: path})
		case event, ok := <-fs.Events:
			if !ok {
				return nil
			}
			if path == "" {
				continue
			}
			relevant := filepath.Clean(event.Name) == path
			if filepath.Clean(event.Name) == directory && event.Op&(fsnotify.Rename|fsnotify.Remove) != 0 {
				_ = fs.Remove(directory)
				directory = ""
				relevant = true
			}
			if directory != filepath.Dir(path) {
				moved, err := refresh()
				if err != nil {
					emit(Error{fmt.Errorf("restore source watch: %w", err)})
				}
				if moved {
					if _, err := os.Stat(path); err == nil {
						relevant = true
					}
				}
			}
			if !relevant || event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}

			stopTimer()
			timer.Reset(200 * time.Millisecond)
			ticks = timer.C
		case <-ticks:
			ticks = nil
			emit(Changed{Path: path})
		case err, ok := <-fs.Errors:
			if !ok {
				return nil
			}
			emit(Error{fmt.Errorf("filesystem watcher: %w", err)})
		}
	}
}
