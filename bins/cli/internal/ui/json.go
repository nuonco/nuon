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
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}

	cliUserErr := &CLIUserError{}
	if errors.As(err, &cliUserErr) {
		PrintJSON(jsonError{
			Error: err.Error(),
		})
		return err
	}

	userErr, ok := nuon.ToUserError(err)
	if ok {
		PrintJSON(userErr)
		return err
	}

	if nuon.IsServerError(err) {
		PrintJSON(jsonError{
			Error: defaultServerErrorMessage,
		})
		return err
	}

	PrintJSON(jsonError{
		Error: defaultUnknownErrorMessage,
	})
	return err
}
