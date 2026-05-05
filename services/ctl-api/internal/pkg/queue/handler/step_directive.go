package handler

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// extractStepDirective inspects err for an stderr.ErrUser and returns the
// pieces the queue handler needs to persist a typed step failure:
//
//   - humanDesc: ErrUser.Description when present (else "" — caller keeps
//     its existing fallback to signal.HumanError).
//   - meta: the metadata fields to merge into the QueueSignal status
//     metadata so the directive/code/fields persist across the DB boundary.
//   - payload: the StepErrorPayload to attach as ApplicationError details
//     on the temporal NonRetryableApplicationError. Caller should check
//     payload.IsZero() before attaching.
//
// When err is not an stderr.ErrUser, all return values are zero / empty.
func extractStepDirective(err error) (humanDesc string, meta map[string]any, payload stderr.StepErrorPayload) {
	u, ok := stderr.IsUserError(err)
	if !ok {
		return "", nil, stderr.StepErrorPayload{}
	}
	return u.Description, stderr.MetadataFromErrUser(u), stderr.PayloadFromErrUser(u)
}
