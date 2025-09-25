package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// for text or ui elements
	PrimaryColor   = lipgloss.Color("63") // Bright Purple
	SecondaryColor = lipgloss.Color("14") // Bright Cyan
	AccentColor    = lipgloss.Color("11") // Bright Yellow
	Dim            = lipgloss.Color("97") // Dark Purple

	// text colors
	TextColor   = lipgloss.Color("")  // Default foreground
	SubtleColor = lipgloss.Color("8") // Bright Black (typically gray)

	BorderColor      = lipgloss.Color("8") // Bright Black (typically gray)
	BorderFocusColor = lipgloss.Color("8") // A purple (typically gray)
	BorderBlurColor  = lipgloss.Color("8") // if not the default, a dark gray (typically gray)

	// very close to light/dark background so it is barely visible
	Ghost = lipgloss.AdaptiveColor{Light: "93", Dark: "17"} //
)
