package ui

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cockroachdb/errors/withstack"

	"github.com/nuonco/nuon/pkg/errs"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

func PrintJSON(data interface{}) {
	if agentEnabled() {
		emitAgentSuccess(data)
		return
	}
	j, _ := json.Marshal(data)
	fmt.Println(string(j))
}

type jsonError struct {
	Error string `json:"error"`
}

// PrintJSONError renders an error in the active output mode. It is equivalent to
// PrintError: agent -> envelope, json -> JSON error object, table -> human text.
// Both entry points exist for historical call sites; prefer either.
func PrintJSONError(err error) error {
	if agentEnabled() {
		return emitAgentError(err)
	}
	if !jsonOutputEnabled() {
		return printHumanError(err)
	}
	return emitJSONError(err)
}

func emitJSONError(err error) error {
	// Construct a stack trace if this error doesn't already have one
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}

	// API user errors keep their structured shape so JSON consumers can read the
	// individual fields rather than a flattened string.
	cliUserErr := &CLIUserError{}
	if !errors.As(err, &cliUserErr) {
		if userErr, ok := nuon.ToUserError(err); ok {
			PrintJSON(userErr)
			return err
		}
	}

	PrintJSON(jsonError{
		Error: resolveError(err).msg,
	})
	return err
}
