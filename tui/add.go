package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/veryshyjelly/eitri/core"
	"github.com/veryshyjelly/eitri/watcher"
)

// AddTestCaseRequest is processed by the UI even while an overlay is open.
// Reply must have capacity one so acknowledgment never blocks the event loop.
// A canceled request is rejected before changing the active problem's tests.
type AddTestCaseRequest struct {
	Context context.Context
	Test    core.Test
	Reply   chan<- error
}

func (m Model) addTestCase(test core.Test) Model {
	if m.runCancel != nil {
		m.runCancel()
	}
	m.generation++
	m.Status = core.NotAvailable
	m.Tests = append(m.Tests, core.Testcase{
		Input: test.Input, Answer: test.Output, Status: core.NotAvailable,
	})
	return m.updatePanes()
}

type AddTestCase struct {
	parent Model
	form   huh.Model

	input  string
	output string
}

func NewAddTestCase(parent Model) *AddTestCase {
	t := &AddTestCase{parent: parent}

	t.form = huh.NewForm(
		huh.NewGroup(
			huh.NewText().Title("Stdin").Value(&t.input),
			huh.NewText().Title("Stdout").Value(&t.output),
		),
	)

	return t
}

func (m *AddTestCase) Init() tea.Cmd {
	return m.form.Init()
}

func (m *AddTestCase) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		return m.parent.updatePanes(), nil
	}

	// Background messages still belong to the underlying problem model.
	switch msg.(type) {
	case core.Info:
		return m.parent.Update(msg)
	case AddTestCaseRequest, runResult, prepared, copyResult, clipboardResult, operationError, watcher.Changed, watcher.Error, tea.WindowSizeMsg:
		parent, cmd := m.parent.Update(msg)
		m.parent = parent.(Model)
		return m, cmd
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)

	if m.form.(*huh.Form).State == huh.StateCompleted {
		return m.parent.addTestCase(core.Test{Input: m.input, Output: m.output}), nil
	}

	return m, cmd
}

func (m *AddTestCase) View() tea.View {
	v := tea.NewView(m.form.View())
	v.AltScreen = true
	return v
}
