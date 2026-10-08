package pager

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestPagerScrollsAndQuits(t *testing.T) {
	content := strings.Repeat("line\n", 40)
	m := model{content: strings.TrimRight(content, "\n")}

	next, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	m = next.(model)
	if !m.ready {
		t.Fatal("pager did not size itself")
	}
	if m.viewport.YOffset() != 0 {
		t.Fatalf("offset = %d, want 0", m.viewport.YOffset())
	}

	next, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	m = next.(model)
	if m.viewport.YOffset() != 1 {
		t.Fatalf("offset after j = %d, want 1", m.viewport.YOffset())
	}

	next, _ = m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	m = next.(model)
	if m.viewport.YOffset() <= 1 {
		t.Fatalf("G did not move to the end, offset %d", m.viewport.YOffset())
	}

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatal("q did not quit")
	}
}

func TestLineCount(t *testing.T) {
	if got := lineCount("a\nb\n"); got != 2 {
		t.Fatalf("lineCount = %d, want 2", got)
	}
}
