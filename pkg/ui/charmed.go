/*
1. Charmed version of the Step method
2. PrintEnv with redaction list
*/
package ui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
var checkMark = lipgloss.NewStyle().SetString("✓").
	Foreground(special).
	PaddingRight(1).
	String()

var shellSymbol = lipgloss.NewStyle().SetString("$").
	Foreground(special).
	PaddingLeft(2).
	PaddingRight(1).
	String()

func (l *logger) Step2(msg string, args ...any) {
	if l.JSON {
		l.Zap.Info(fmt.Sprintf(msg, args...))
		return
	}

	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func Step2(ctx context.Context, msg string, args ...any) {
	log, err := FromContext(ctx)
	if err != nil {
		return
	}
	s := checkMark + lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EEEEEE")).
		Render(msg)
	log.Step2(s, args...)
}

func PrintEnv(ctx context.Context, env map[string]string) {
	redacted := []string{
		"api_token",
		"ca_data",
		"token",
		"aws_session",
		"secret_key",
		"secret",
		"aws_access",
		"github_app_key",
	}
	maxKLen := 0
	for k, _ := range env {
		if len(k) > maxKLen {
			maxKLen = len(k)
		}
	}

	for k, v := range env {
		redact := false
		vp := v
		for _, rsub := range redacted {
			if strings.Contains(strings.ToLower(k), strings.ToLower(rsub)) {
				vp = fmt.Sprintf("▉▉▉▉▉▉ [REDACTED %d chars]", len(v))
				redact = true
			}
		}
		s := shellSymbol
		s += k + strings.Repeat(" ", maxKLen-len(k)) + " = "
		if redact {
			s += lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).Render(vp)
		} else {
			s += vp
		}
		fmt.Fprint(os.Stderr, s+"\n")
	}
}
