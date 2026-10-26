package tui

import (
	"context"
	"sync"

	tea "charm.land/bubbletea/v2"
)

// tasks bounds background work to a single operation. New run requests cancel
// the previous context; commands queued by Bubble Tea check it before doing IO.
// Close excludes new work and joins all operations already admitted.
type tasks struct {
	ctx    context.Context
	cancel context.CancelFunc
	gate   chan struct{}
	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
}

func newTasks(parent context.Context) *tasks {
	ctx, cancel := context.WithCancel(parent)
	return &tasks{ctx: ctx, cancel: cancel, gate: make(chan struct{}, 1)}
}
func (t *tasks) do(ctx context.Context, work func(context.Context) tea.Msg) tea.Msg {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.wg.Add(1)
	t.mu.Unlock()
	defer t.wg.Done()
	select {
	case t.gate <- struct{}{}:
	case <-ctx.Done():
		return nil
	}
	defer func() { <-t.gate }()
	if ctx.Err() != nil {
		return nil
	}
	return work(ctx)
}
func (t *tasks) close() {
	t.mu.Lock()
	t.closed = true
	t.cancel()
	t.mu.Unlock()
	t.wg.Wait()
}
