package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectCommandsPreserveQuotedArgumentsWithoutEvaluatingShellExpressions(t *testing.T) {
	for _, tc := range []struct {
		command string
		want    []string
	}{
		{`/dnsutils`, []string{"/dnsutils"}},
		{`/tool --label 'two words' "" "literal*"`, []string{"/tool", "--label", "two words", "", "literal*"}},
		{`/tool '$(touch /tmp/unexpected)'`, []string{"/tool", "$(touch /tmp/unexpected)"}},
		{`/tool "$TOKEN"`, nil},
		{`/tool $(other-command)`, nil},
		{`/tool *.json`, nil},
		{`/tool ~root`, nil},
		{`/tool > result.json`, nil},
		{`/tool | other-command`, nil},
		{`/tool && other-command`, nil},
		{`/tool; other-command`, nil},
		{`/tool &`, nil},
		{`! /tool`, nil},
		{`NAME=value /tool`, nil},
		{`echo hello`, nil},
		{`./script.sh`, nil},
		{`/tool "unterminated`, nil},
	} {
		t.Run(tc.command, func(t *testing.T) { assert.Equal(t, tc.want, directContainerCommand(tc.command)) })
	}
}
