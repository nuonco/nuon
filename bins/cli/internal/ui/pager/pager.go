// Package pager shows text that is taller than the terminal in an alt-screen
// Bubble Tea viewport. Callers that are not a TTY should print the text.
package pager

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"golang.org/x/term"

	"github.com/nuonco/nuon/bins/cli/internal/ui/teaprogram"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

// NeedsPager reports whether content is taller than the terminal. A non-TTY
// never pages, so pipes and agents keep the full text.
func NeedsPager(content string) bool {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		return false
	}
	_, height, err := term.GetSize(fd)
	if err != nil || height <= 1 {
		return false
	}
	return lineCount(content) > height-1
}

// Run displays content in an alt-screen pager. q, esc, and ctrl+c quit.
func Run(content string) error {
	p := teaprogram.NewProgram(model{content: strings.TrimRight(content, "\n")})
	_, err := p.Run()
	return err
}

type model struct {
	content  string
	viewport viewport.Model
	ready    bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		height := max(msg.Height-1, 1)
		width := max(msg.Width, 1)
		if !m.ready {
			m.viewport = viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
			m.viewport.SoftWrap = true
			m.viewport.SetContent(m.content)
			m.ready = true
			return m, nil
		}
		m.viewport.SetWidth(width)
		m.viewport.SetHeight(height)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "g", "home":
			m.viewport.SetYOffset(0)
			return m, nil
		case "G", "end":
			m.viewport.SetYOffset(m.viewport.TotalLineCount())
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(m.view())
	v.AltScreen = true
	return v
}

func (m model) view() string {
	if !m.ready {
		return ""
	}
	percent := int(m.viewport.ScrollPercent() * 100)
	footer := styles.TextDim.Render(fmt.Sprintf(" %d%%   ↑/k  ↓/j  space  g/G  q quit", percent))
	return lipgloss.JoinVertical(lipgloss.Top, m.viewport.View(), footer)
}

func lineCount(content string) int {
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}
