package ui

import (
	"context"
	"fmt"
	"os"

	"github.com/kyokomi/emoji"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
	"go.uber.org/zap"
)

const (
	ColorGreen  = "\033[032m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

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
	fmt.Fprintf(os.Stderr, message, args...)
}

// Error records an error that happened, with a "x"
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

	emoji.Fprintf(os.Stderr, ":siren: %s\n", err.Error())
}
