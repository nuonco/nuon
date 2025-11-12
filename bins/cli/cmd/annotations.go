package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
)

const skipAuthAnnotationKey string = "skip_auth"

func skipAuthAnnotation() map[string]string {
	return map[string]string{
		skipAuthAnnotationKey: strconv.FormatBool(true),
	}
}

func hasSkipAuthAnnotation(cmd *cobra.Command) bool {
	skipAuth, ok := cmd.Annotations[skipAuthAnnotationKey]
	if !ok {
		return false
	}

	return skipAuth == strconv.FormatBool(true)
}
