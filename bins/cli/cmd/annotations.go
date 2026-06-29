package cmd

import (
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const skipAuthAnnotationKey string = "skip_auth"
const tuiAnnotationKey string = "tui"
const previewAnnotationKey string = "preview"
const outputsAnnotationKey string = "outputs"

// TUI annotation values
const (
	TUIAltScreen  = "alt-screen"
	TUIContextual = "contextual"
)

// Output annotation values, matching --output formats.
const (
	OutputTable = "table"
	OutputJSON  = "json"
	OutputAgent = "agent"
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

// outputsAnnotation declares which --output formats a command supports.
// Commands without it support all formats (table, json, agent).
func outputsAnnotation(types ...string) map[string]string {
	return map[string]string{
		outputsAnnotationKey: strings.Join(types, ","),
	}
}

func supportedOutputs(cmd *cobra.Command) []string {
	v, ok := cmd.Annotations[outputsAnnotationKey]
	if !ok || v == "" {
		return []string{OutputTable, OutputJSON, OutputAgent}
	}
	return strings.Split(v, ",")
}

func supportsOutput(cmd *cobra.Command, out string) bool {
	return slices.Contains(supportedOutputs(cmd), out)
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
