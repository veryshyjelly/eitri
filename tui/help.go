package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
)

type Help struct {
	parent tea.Model

	height int
	width  int
}

func (h Help) Init() tea.Cmd {
	return nil
}

func (h Help) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Debug("Key pressed", "key", msg.String())
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit, DefaultKeyMap.Help):
			return h.parent, nil
		}
	case tea.WindowSizeMsg:
		h.parent, cmd = h.parent.Update(msg)
		h.width = max(1, msg.Width-2)
		h.height = max(1, msg.Height-2)
	default:
		h.parent, cmd = h.parent.Update(msg)
	}
	return h, cmd
}

// View renders the full-screen help interface displaying all available
// keyboard shortcuts. The view is dynamically centered within the current
// terminal dimensions.
func (h Help) View() tea.View {
	hi := help.New()
	hi.ShowAll = true
	hi.SetWidth(h.width)

	s := containerStyle.
		Height(h.height + 2).
		Width(h.width + 2).
		Render(
			lipgloss.Place(
				h.width,
				h.height,
				lipgloss.Center,
				lipgloss.Center,
				hi.View(DefaultKeyMap),
			),
		)

	v := tea.NewView(s)
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	return v
}
