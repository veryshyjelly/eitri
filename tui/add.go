package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/veryshyjelly/eitri/core"
	"github.com/veryshyjelly/eitri/watcher"
)

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
	case runResult, prepared, copyResult, clipboardResult, operationError, watcher.Changed, watcher.Error, tea.WindowSizeMsg:
		parent, cmd := m.parent.Update(msg)
		m.parent = parent.(Model)
		return m, cmd
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)

	if m.form.(*huh.Form).State == huh.StateCompleted {
		if m.parent.runCancel != nil {
			m.parent.runCancel()
		}
		m.parent.generation++
		m.parent.Status = core.NotAvailable
		m.parent.Tests = append(m.parent.Tests,
			core.Testcase{
				Input:  m.input,
				Answer: m.output,
				Status: core.NotAvailable,
			},
		)
		return m.parent.updatePanes(), nil
	}

	return m, cmd
}

func (m *AddTestCase) View() tea.View {
	v := tea.NewView(m.form.View())
	v.AltScreen = true
	return v
}
