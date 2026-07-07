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
