package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

type agentEnvelope struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error *agentError `json:"error,omitempty"`
}

type agentError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func agentEnabled() bool { return agentmode.Enabled() }

// jsonOutput is set when --output json is selected (non-agent). It makes the
// error helpers emit a JSON error object instead of human-styled text.
var jsonOutput bool

// SetJSONOutput toggles plain JSON output mode. Called once during output
// resolution. Agent mode is tracked separately via agentmode and takes
// precedence in every helper.
func SetJSONOutput(v bool) { jsonOutput = v }

func jsonOutputEnabled() bool { return jsonOutput }

func emitAgentSuccess(data interface{}) {
	emitAgent(agentEnvelope{OK: true, Data: data})
}

func emitAgentError(err error) error {
	code, msg := classifyError(err)
	emitAgent(agentEnvelope{Error: &agentError{Code: code, Message: msg}})
	return err
}

func emitAgent(env agentEnvelope) {
	j, mErr := json.Marshal(env)
	if mErr != nil {
		j, _ = json.Marshal(agentEnvelope{Error: &agentError{Code: "error", Message: mErr.Error()}})
	}
	fmt.Fprintln(os.Stdout, string(j))
}

// classifyError maps an error to a stable machine code and a human message.
func classifyError(err error) (string, string) {
	var exitErr *ErrExitCode
	if errors.As(err, &exitErr) && exitErr.Code != "" {
		return exitErr.Code, exitErr.Error()
	}

	cliUserErr := &CLIUserError{}
	if errors.As(err, &cliUserErr) {
		return "user_error", cliUserErr.Msg
	}

	code := ""
	switch {
	case nuon.IsNotFound(err):
		code = "not_found"
	case nuon.IsUnauthorized(err):
		code = "unauthorized"
	case nuon.IsForbidden(err):
		code = "forbidden"
	case nuon.IsBadRequest(err):
		code = "invalid_request"
	case nuon.IsServerError(err):
		return "server_error", defaultServerErrorMessage
	}

	if msg, ok := nuon.ToAPIError(err); ok {
		if code == "" {
			code = "api_error"
		}
		return code, msg
	}
	if code == "" {
		code = "error"
	}
	return code, err.Error()
}
