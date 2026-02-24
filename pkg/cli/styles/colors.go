package styles

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

var hasDarkBG = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
var lightDark = lipgloss.LightDark(hasDarkBG)

// Official Colors
// resolved at init time via lightDark
var (
	PrimaryColor        = lightDark(lipgloss.Color("#8040BF"), lipgloss.Color("#D6B0FC"))
	SecondaryColor      = lightDark(lipgloss.Color("#527FE8"), lipgloss.Color("#99B7FF"))
	AccentColor         = lightDark(lipgloss.Color("#D6B0FC"), lipgloss.Color("#8040BF"))
	TextColor           = lightDark(lipgloss.Color("#000000"), lipgloss.Color("#FFFFFF"))
	SubtleColor         = lightDark(lipgloss.Color("#C3C3C3"), lipgloss.Color("#B9B9B9"))
	SuccessColor        = lightDark(lipgloss.Color("#439B92"), lipgloss.Color("#5BBFB5"))
	WarningColor        = lightDark(lipgloss.Color("#FCA04A"), lipgloss.Color("#FFBD7F"))
	ErrorColor          = lightDark(lipgloss.Color("#991B1B"), lipgloss.Color("#FF8383"))
	InfoColor           = lightDark(lipgloss.Color("#527FE8"), lipgloss.Color("#527FE8"))
	BorderActiveColor   = lightDark(lipgloss.Color("#8040BF"), lipgloss.Color("#8040BF"))
	BorderInactiveColor = lightDark(lipgloss.Color("#C3C3C3"), lipgloss.Color("#4F4F4F"))
	PrimaryBGColor      = lightDark(lipgloss.Color("#F8F6F6"), lipgloss.Color("#1B242C"))

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
	Dim   color.Color = lipgloss.Color("#d149b7")
	Ghost color.Color = lightDark(lipgloss.Color("#93"), lipgloss.Color("#17"))
)
