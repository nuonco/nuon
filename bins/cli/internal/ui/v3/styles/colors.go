package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Official Colors
// these are internal values
var (
	lightPrimaryColor        = lipgloss.CompleteColor{TrueColor: "#8040BF", ANSI256: "99", ANSI: "5"}
	lightSecondaryColor      = lipgloss.CompleteColor{TrueColor: "#527FE8", ANSI256: "69", ANSI: "12"}
	lightAccentColor         = lipgloss.CompleteColor{TrueColor: "#D6B0FC", ANSI256: "183", ANSI: "13"}
	lightTextColor           = lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"}
	lightSubtleColor         = lipgloss.CompleteColor{TrueColor: "#C3C3C3", ANSI256: "250", ANSI: "7"}
	lightSuccessColor        = lipgloss.CompleteColor{TrueColor: "#439B92", ANSI256: "72", ANSI: "10"}
	lightWarningColor        = lipgloss.CompleteColor{TrueColor: "#FCA04A", ANSI256: "214", ANSI: "11"}
	lightErrorColor          = lipgloss.CompleteColor{TrueColor: "#991B1B", ANSI256: "88", ANSI: "1"}
	lightInfoColor           = lipgloss.CompleteColor{TrueColor: "#527FE8", ANSI256: "69", ANSI: "12"}
	lightBorderActiveColor   = lipgloss.CompleteColor{TrueColor: "#8040BF", ANSI256: "99", ANSI: "5"}
	lightBorderInactiveColor = lipgloss.CompleteColor{TrueColor: "#C3C3C3", ANSI256: "250", ANSI: "7"}
	lightPrimaryBGColor      = lipgloss.CompleteColor{TrueColor: "#F8F6F6", ANSI256: "255", ANSI: "15"}

	// Dark
	darkPrimaryColor        = lipgloss.CompleteColor{TrueColor: "#D6B0FC", ANSI256: "183", ANSI: "13"}
	darkSecondaryColor      = lipgloss.CompleteColor{TrueColor: "#99B7FF", ANSI256: "111", ANSI: "12"}
	darkAccentColor         = lipgloss.CompleteColor{TrueColor: "#8040BF", ANSI256: "99", ANSI: "5"}
	darkTextColor           = lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "15", ANSI: "5"}
	darkSubtleColor         = lipgloss.CompleteColor{TrueColor: "#B9B9B9", ANSI256: "249", ANSI: "7"}
	darkSuccessColor        = lipgloss.CompleteColor{TrueColor: "#5BBFB5", ANSI256: "80", ANSI: "10"}
	darkWarningColor        = lipgloss.CompleteColor{TrueColor: "#FFBD7F", ANSI256: "223", ANSI: "11"}
	darkErrorColor          = lipgloss.CompleteColor{TrueColor: "#FF8383", ANSI256: "210", ANSI: "9"}
	darkInfoColor           = lipgloss.CompleteColor{TrueColor: "#527FE8", ANSI256: "69", ANSI: "12"}
	darkBorderActiveColor   = lipgloss.CompleteColor{TrueColor: "#8040BF", ANSI256: "99", ANSI: "5"}
	darkBorderInactiveColor = lipgloss.CompleteColor{TrueColor: "#4F4F4F", ANSI256: "238", ANSI: "8"}
	darkPrimaryBGColor      = lipgloss.CompleteColor{TrueColor: "#1B242C", ANSI256: "235", ANSI: "0"}
)

// Official Complete Adaptive Colors
// these are what we want to export
var (
	PrimaryColor        = lipgloss.CompleteAdaptiveColor{Light: lightPrimaryColor, Dark: darkPrimaryColor}
	SecondaryColor      = lipgloss.CompleteAdaptiveColor{Light: lightSecondaryColor, Dark: darkSecondaryColor}
	AccentColor         = lipgloss.CompleteAdaptiveColor{Light: lightAccentColor, Dark: darkAccentColor}
	TextColor           = lipgloss.CompleteAdaptiveColor{Light: lightTextColor, Dark: darkTextColor}
	SubtleColor         = lipgloss.CompleteAdaptiveColor{Light: lightSubtleColor, Dark: darkSubtleColor}
	SuccessColor        = lipgloss.CompleteAdaptiveColor{Light: lightSuccessColor, Dark: darkSuccessColor}
	WarningColor        = lipgloss.CompleteAdaptiveColor{Light: lightWarningColor, Dark: darkWarningColor}
	ErrorColor          = lipgloss.CompleteAdaptiveColor{Light: lightErrorColor, Dark: darkErrorColor}
	InfoColor           = lipgloss.CompleteAdaptiveColor{Light: lightInfoColor, Dark: darkInfoColor}
	BorderActiveColor   = lipgloss.CompleteAdaptiveColor{Light: lightBorderActiveColor, Dark: darkBorderActiveColor}
	BorderInactiveColor = lipgloss.CompleteAdaptiveColor{Light: lightBorderInactiveColor, Dark: darkBorderInactiveColor}
	PrimaryBGColor      = lipgloss.CompleteAdaptiveColor{Light: lightPrimaryBGColor, Dark: darkPrimaryBGColor}

	// Text
	TextPrimary   = lipgloss.NewStyle().Foreground(PrimaryColor)
	TextSecondary = lipgloss.NewStyle().Foreground(SecondaryColor)
	TextAccent    = lipgloss.NewStyle().Foreground(AccentColor)
	TextDefault   = lipgloss.NewStyle().Foreground(TextColor)
	TextSubtle    = lipgloss.NewStyle().Foreground(SubtleColor)
	TextSuccess   = lipgloss.NewStyle().Foreground(SuccessColor)
	TextWarning   = lipgloss.NewStyle().Foreground(WarningColor)
	TextError     = lipgloss.NewStyle().Foreground(ErrorColor)
	TextInfo      = lipgloss.NewStyle().Foreground(InfoColor)
)

// holdovers we don't know how to get rid of yet
var (
	// we do want to keep thise and need to integrate them
	Dim = lipgloss.CompleteAdaptiveColor{
		Light: lipgloss.CompleteColor{
			TrueColor: "#d149b7",
			ANSI256:   "97",
			ANSI:      "5",
		},
		Dark: lipgloss.CompleteColor{
			TrueColor: "#d149b7",
			ANSI256:   "97",
			ANSI:      "5",
		},
	}
	Ghost = lipgloss.AdaptiveColor{Light: "93", Dark: "17"} //
)
