package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/veryshyjelly/eitri/core"
)

// Splash is used to display the landing page until we get our first info
type Splash struct {
	home tea.Model

	height int
	width  int
}

func NewSplash(home tea.Model) Splash {
	return Splash{home: home}
}

func (s Splash) Init() tea.Cmd {
	return nil
}

func (s Splash) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			return s, tea.Quit
		case key.Matches(msg, DefaultKeyMap.Help):
			h := Help{parent: s, height: s.height, width: s.width}
			return h, nil
		}
	case tea.WindowSizeMsg:
		s.width = max(1, msg.Width-2)
		s.height = max(1, msg.Height-2)
		s.home, cmd = s.home.Update(msg)
	case core.Info:
		return s.home.Update(msg)
	}
	return s, cmd
}

// View renders the idle state UI, displaying a prompt instructing
// the user to select a problem via the browser extension, alongside the compact
// help menu at the bottom of the screen.
func (s Splash) View() tea.View {
	mess := "Select problem from competitive companion"
	content := lipgloss.NewLayer(lipgloss.Place(
		s.width,
		s.height,
		lipgloss.Center,
		lipgloss.Center,
		waitMessageStyle.Render(mess),
	))

	h := help.New()
	helpLayer := lipgloss.NewLayer(
		lipgloss.PlaceHorizontal(s.width, lipgloss.Center,
			h.View(DefaultKeyMap)),
	).Y(s.height - 1)

	c := containerStyle.
		Height(s.height + 2).
		Width(s.width + 2).
		Render(lipgloss.NewCompositor(content, helpLayer).Render())

	v := tea.NewView(c)
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	return v
}
