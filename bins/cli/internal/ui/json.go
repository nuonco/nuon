package ui

import (
	"encoding/json"
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

// PrintIndentedJSON renders data as indented JSON to the mode-aware human writer.
// Used for the table (human) path of commands whose payload is an arbitrary
// config blob with no natural tabular form.
func PrintIndentedJSON(data interface{}) {
	j, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		Println(err.Error())
		return
	}
	Println(string(j))
}

type jsonError struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// PrintJSONError renders an error in the active output mode. It is equivalent
// to PrintError; both entry points exist for historical call sites.
func PrintJSONError(err error) error {
	return PrintError(err)
}

func emitJSONError(err error) error {
	// Construct a stack trace if this error doesn't already have one
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}

	// API user errors keep their historical shape: the full error object.
	if userErr, ok := nuon.ToUserError(err); ok {
		PrintJSON(userErr)
		return err
	}

	// Everything else classifies exactly like the agent envelope, so json and
	// agent callers see the same message for the same failure. Messages that
	// would leak internal details fall back to the generic one, matching the
	// table renderer's policy.
	code, msg := classifyError(err)
	if containsTechnicalError(msg) {
		msg = defaultUnknownErrorMessage
	}
	PrintJSON(jsonError{Error: msg, Code: code})
	return err
}
