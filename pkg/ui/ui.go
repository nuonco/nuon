package ui

import (
	"context"
	"fmt"
	"os"

	"github.com/kyokomi/emoji"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/cli/styles"
)

const (
	ColorGreen  = "\033[032m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

var quietMode bool

func SetQuietMode(val bool) {
	quietMode = val
}

func GetStatusColor(status string) string {
	var statusColor string

	switch status {
	case "active":
		statusColor = ColorGreen
	case "failed":
		statusColor = ColorRed
	default:
		statusColor = ColorYellow
	}

	return statusColor
}

func Line(ctx context.Context, msg string, args ...any) {
	log, err := FromContext(ctx)
	if err != nil {
		return
	}

	log.Step(msg, args...)
}

// Step records a step that happened, with a "check"
func Step(ctx context.Context, msg string, args ...any) {
	log, err := FromContext(ctx)
	if err != nil {
		return
	}

	log.Step(msg, args...)
}

func (l *logger) Step(msg string, args ...any) {
	if l.JSON {
		l.Zap.Info(fmt.Sprintf(msg, args...))
		return
	}
	message := fmt.Sprintf("%s %s\n", styles.TextSuccess.Render("✔"), msg)

	if !quietMode {
		fmt.Fprintf(os.Stderr, message, args...)
	}

	writeToLogFile(message, args...)
}

func Error(ctx context.Context, userErr error) {
	log, err := FromContext(ctx)
	if err != nil {
		return
	}

	log.Error(userErr)
}

func (l *logger) Error(err error) {
	if l.JSON {
		l.Zap.Error("error", zap.Error(err))
		return
	}

	msg := emoji.Sprintf(":siren: %s\n", err.Error())
	emoji.Fprintf(os.Stderr, msg)
	writeToLogFile(msg)
}

func EnvVars(envVars map[string]any) {
	if len(envVars) == 0 {
		return
	}

	maxKeyLen := 0
	for k := range envVars {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	dollarStyle := styles.TextDim
	keyStyle := styles.TextSuccess
	eqStyle := styles.TextDim
	valStyle := styles.TextSubtle

	for k, v := range envVars {
		valStr := fmt.Sprintf("%v", v)
		if len(valStr) > 16 {
			valStr = valStr[:6] + "..." + valStr[len(valStr)-7:]
		}

		paddedKey := fmt.Sprintf("%-*s", maxKeyLen, k)
		fmt.Fprintf(os.Stderr, "  %s %s %s %s\n",
			dollarStyle.Render("$"),
			keyStyle.Render(paddedKey),
			eqStyle.Render("="),
			valStyle.Render(valStr),
		)
	}
}
