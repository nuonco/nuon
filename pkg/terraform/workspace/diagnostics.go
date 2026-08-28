package workspace

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
)

// extractDiagnostics pulls error diagnostics out of a terraform -json output
// stream so plan/apply failures surface the underlying terraform error
// instead of just "exit status 1".
func extractDiagnostics(raw []byte) string {
	var msgs []string
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var entry struct {
			Level      string `json:"@level"`
			Message    string `json:"@message"`
			Diagnostic struct {
				Summary string `json:"summary"`
				Detail  string `json:"detail"`
			} `json:"diagnostic"`
		}
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if entry.Level != "error" {
			continue
		}
		msg := entry.Message
		if entry.Diagnostic.Detail != "" {
			msg += ": " + entry.Diagnostic.Detail
		}
		if msg != "" {
			msgs = append(msgs, msg)
		}
	}
	return strings.Join(msgs, "; ")
}
