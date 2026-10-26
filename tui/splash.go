package tui

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/veryshyjelly/eitri/core"
)

const splashWordmark = `███████╗██╗████████╗██████╗ ██╗
██╔════╝██║╚══██╔══╝██╔══██╗██║
█████╗  ██║   ██║   ██████╔╝██║
██╔══╝  ██║   ██║   ██╔══██╗██║
███████╗██║   ██║   ██║  ██║██║
╚══════╝╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝`

var (
	splashLogo        = renderSplashLogo()
	splashPromptStyle = waitMessageStyle.Foreground(lipgloss.Color("#89DCEB"))
	splashHintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7F849C"))
)

// Render the text gradient once; resizing only recomposes the dashboard.
func renderSplashLogo() string {
	palette := [...]color.RGBA{
		{R: 0xFA, G: 0xB3, B: 0x87, A: 0xFF},
		{R: 0xF5, G: 0xC2, B: 0xE7, A: 0xFF},
		{R: 0xCB, G: 0xA6, B: 0xF7, A: 0xFF},
	}
	width := lipgloss.Width(splashWordmark)
	styles := make([]lipgloss.Style, width)
	for column := range styles {
		position := float64(column) * 2 / float64(width-1)
		segment := min(int(position), 1)
		fraction := position - float64(segment)
		from, to := palette[segment], palette[segment+1]
		blend := func(a, b uint8) uint8 {
			return uint8(float64(a)*(1-fraction) + float64(b)*fraction)
		}
		styles[column] = lipgloss.NewStyle().Foreground(color.RGBA{
			R: blend(from.R, to.R), G: blend(from.G, to.G), B: blend(from.B, to.B), A: 0xFF,
		})
	}
	lines := strings.Split(splashWordmark, "\n")
	for row, line := range lines {
		var rendered strings.Builder
		for column, glyph := range []rune(line) {
			rendered.WriteString(styles[column].Render(string(glyph)))
		}
		lines[row] = rendered.String()
	}
	return strings.Join(lines, "\n")
}

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
			h := Help{parent: s, height: max(1, s.height-2), width: max(1, s.width-2)}
			return h, nil
		}
	case tea.WindowSizeMsg:
		s.width = max(1, msg.Width)
		s.height = max(1, msg.Height)
		s.home, cmd = s.home.Update(msg)
	case core.Info:
		return s.home.Update(msg)
	case AddTestCaseRequest:
		s.home, cmd = s.home.Update(msg)
	}
	return s, cmd
}

// View centers the wordmark, Companion prompt, and shortcuts. Small terminals
// use a compact title or just the prompt instead of clipping the ASCII artwork.
func (s Splash) View() tea.View {
	terminalWidth, terminalHeight := max(1, s.width), max(1, s.height)
	width, height := max(1, terminalWidth-2), max(1, terminalHeight-2)
	prompt := splashPromptStyle.Width(width).Align(lipgloss.Center).
		Render("Select a problem from Competitive Companion")
	hints := splashHintStyle.Width(width).Align(lipgloss.Center).
		Render("? help    ·    q quit")
	title := lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Render("EITRI")
	dashboard := prompt
	for _, candidate := range []string{
		lipgloss.JoinVertical(lipgloss.Center, splashLogo, "", prompt, "", hints),
		lipgloss.JoinVertical(lipgloss.Center, title, "", prompt, "", hints),
		lipgloss.JoinVertical(lipgloss.Center, prompt, "", hints),
	} {
		if lipgloss.Width(candidate) <= width && lipgloss.Height(candidate) <= height {
			dashboard = candidate
			break
		}
	}
	dashboard = lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(dashboard)

	c := containerStyle.
		Height(height + 2).
		Width(width + 2).
		Render(lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dashboard))
	if terminalWidth < 3 || terminalHeight < 3 {
		c = lipgloss.NewStyle().MaxWidth(terminalWidth).MaxHeight(terminalHeight).Render("EITRI")
	}

	v := tea.NewView(c)
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	return v
}
