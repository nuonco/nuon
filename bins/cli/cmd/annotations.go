package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
)

const skipAuthAnnotationKey string = "skip_auth"
const tuiAnnotationKey string = "tui"
const previewAnnotationKey string = "preview"

// TUI annotation values
const (
	TUIAltScreen  = "alt-screen"
	TUIContextual = "contextual"
)

func skipAuthAnnotation() map[string]string {
	return map[string]string{
		skipAuthAnnotationKey: strconv.FormatBool(true),
	}
}

func tuiAnnotation(tuiType string) map[string]string {
	return map[string]string{
		tuiAnnotationKey: tuiType,
	}
}

func previewAnnotation() map[string]string {
	return map[string]string{
		previewAnnotationKey: strconv.FormatBool(true),
	}
}

// annotations merges multiple annotation maps into one.
func annotations(maps ...map[string]string) map[string]string {
	merged := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

func hasSkipAuthAnnotation(cmd *cobra.Command) bool {
	skipAuth, ok := cmd.Annotations[skipAuthAnnotationKey]
	if !ok {
		return false
	}

	return skipAuth == strconv.FormatBool(true)
}
