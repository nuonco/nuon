package ui

import (
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors/withstack"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/pkg/errs"
)

func PrintJSON(data interface{}) {
	j, _ := json.Marshal(data)
	fmt.Println(string(j))
}

type jsonError struct {
	Error string `json:"error"`
}

func PrintJSONError(err error) error {
	// Construct a stack trace if this error doesn't already have one
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
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
