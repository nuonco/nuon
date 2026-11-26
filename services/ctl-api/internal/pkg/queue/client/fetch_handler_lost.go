package client

import (
	stderrors "errors"
	"fmt"
	"strings"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/temporal"
)

const FetchHandlerLostErrType = "FETCH_STEPS_HANDLER_LOST"

func IsFetchHandlerLost(err error) bool {
	var appErr *temporal.ApplicationError
	return stderrors.As(err, &appErr) && appErr.Type() == FetchHandlerLostErrType
}

func handlerLost(err error) bool {
	var notFound *serviceerror.NotFound
	if stderrors.As(err, &notFound) {
		return true
	}
	var appErr *temporal.ApplicationError
	if stderrors.As(err, &appErr) && appErr.Type() == "AcceptedUpdateCompletedWorkflow" {
		return true
	}
	return strings.Contains(err.Error(), "aborted by closing workflow")
}

func handlerLostError(updateName, queueSignalID, runID string, err error) error {
	if !handlerLost(err) {
		return nil
	}
	return temporal.NewNonRetryableApplicationError(
		fmt.Sprintf("%s handler run %s closed before serving signal %s: %s", updateName, runID, queueSignalID, err),
		FetchHandlerLostErrType,
		err,
	)
}
