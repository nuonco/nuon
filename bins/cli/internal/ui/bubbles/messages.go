package bubbles

import "charm.land/lipgloss/v2"

// ColoredText returns text styled with the given hex color.
func ColoredText(text, color string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}

func Green(text string) string {
	return ColoredText(text, "#10b981")
}

func Red(text string) string {
	return ColoredText(text, "#ef4444")
}

func Yellow(text string) string {
	return ColoredText(text, "#f59e0b")
}
