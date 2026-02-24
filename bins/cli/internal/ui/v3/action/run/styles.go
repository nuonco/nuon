package run

import "github.com/nuonco/nuon/pkg/cli/styles"

var (
	appStyle      = styles.Pane
	appStyleBlur  = styles.PaneBlur
	appStyleFocus = styles.PaneFocus
)

// Re-export step style helpers so callers within this package
// don't need a separate import.
var getStepStyle = styles.GetStepStyle
var getStepStatusIcon = styles.GetStepStatusIcon
