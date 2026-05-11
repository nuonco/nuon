package activities

import (
	"context"
	"errors"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

// RecordFromErrorRequest is the JSON-serializable input to the
// RecordFromError activity. It mirrors composite_error.ParseInput but with
// GoErr replaced by GoErrText since errors don't cross activity boundaries.
type RecordFromErrorRequest struct {
	OwnerType string `validate:"required"`
	OwnerID   string `validate:"required"`

	// StepID, when set, instructs the activity to resolve the dispatch
	// ParseContext and InvocationContext from the step's target chain
	// (install_deploy → component + install or install_sandbox_run → install).
	// Resolution is best-effort: if the step's target type is opaque the
	// dispatch falls back to the unknown_error builtin.
	//
	// When set, the resolved fields fill in any zero-valued fields on
	// Invocation; explicit caller values always win on overlap.
	StepID string

	// ParseContext routes the input to matching parsers. When StepID is set
	// and ParseContext is empty, the activity resolves it from the step.
	// Empty string is allowed and falls back to the unknown_error builtin.
	ParseContext composite_error.ParseContext

	Raw      []byte
	ExitCode int

	// GoErrText is the user-facing message from the producing error.
	// Callers must pre-clean it (e.g. via `signal.HumanError`) before
	// dispatching — the parser pipeline does not re-walk Temporal
	// activity / workflow error envelopes. See specs/composite-errors.md
	// §"Lifecycle: write path".
	GoErrText string

	Invocation composite_error.InvocationContext
}

// RecordFromErrorResponse is the JSON-serializable result. The typed
// CompositeError instance does not cross the boundary; callers either
// re-Hydrate through the catalog or — most commonly — read OverrideDirective
// here directly.
type RecordFromErrorResponse struct {
	// ErrorID is the persisted composite_errors row ID.
	ErrorID string

	// Type / Domain / Severity are pulled off the typed instance for
	// convenience (e.g. metrics, log fields). Use ErrorID + Hydrate for the
	// full typed value.
	Type     composite_error.Type
	Domain   composite_error.Domain
	Severity composite_error.Severity

	// OverrideDirective is the merged directive emitted by the typed error
	// (if it implements composite_error.ErrorWithDirective). Empty string
	// means "no opinion — defer to existing signal defaults".
	OverrideDirective string
}

// RecordFromError is the workflow-side entrypoint. It runs the parser
// pipeline against the raw inputs, records the result tree, and returns
// the persisted ErrorID plus any directive override the typed error wants
// to apply.
//
// @temporal-gen-v2 activity
func (a *Activities) RecordFromError(ctx context.Context, req RecordFromErrorRequest) (*RecordFromErrorResponse, error) {
	parseCtx := req.ParseContext
	invocation := req.Invocation

	if req.StepID != "" {
		resolvedCtx, resolvedInv, err := a.helpers.ResolveStepDispatch(ctx, req.StepID)
		if err == nil {
			if parseCtx == "" {
				parseCtx = resolvedCtx
			}
			invocation = mergeInvocation(invocation, resolvedInv)
		}
	}

	in := composite_error.ParseInput{
		Raw:        req.Raw,
		ExitCode:   req.ExitCode,
		Invocation: invocation,
	}
	if req.GoErrText != "" {
		in.GoErr = errors.New(req.GoErrText)
	}

	row, typed, err := a.helpers.RecordFromError(ctx, req.OwnerType, req.OwnerID, parseCtx, in)
	if err != nil {
		return nil, err
	}

	resp := &RecordFromErrorResponse{
		ErrorID:  row.ID,
		Type:     row.Type,
		Domain:   row.Domain,
		Severity: row.Severity,
	}

	if d, ok := typed.(composite_error.ErrorWithDirective); ok {
		resp.OverrideDirective = string(d.OverrideDirective().Kind)
	}
	return resp, nil
}

// mergeInvocation fills empty fields on caller with values from resolved.
// Caller-supplied values always win.
func mergeInvocation(caller, resolved composite_error.InvocationContext) composite_error.InvocationContext {
	mergeIfEmpty(&caller.OrgID, resolved.OrgID)
	mergeIfEmpty(&caller.OwnerID, resolved.OwnerID)
	mergeIfEmpty(&caller.OwnerType, resolved.OwnerType)
	mergeIfEmpty(&caller.ComponentID, resolved.ComponentID)
	mergeIfEmpty(&caller.ComponentType, resolved.ComponentType)
	mergeIfEmpty(&caller.InstallID, resolved.InstallID)
	mergeIfEmpty(&caller.BuildID, resolved.BuildID)
	mergeIfEmpty(&caller.StepID, resolved.StepID)
	mergeIfEmpty(&caller.CloudPlatform, resolved.CloudPlatform)
	return caller
}

// mergeIfEmpty assigns fallback to *dst when *dst is the empty string.
func mergeIfEmpty(dst *string, fallback string) {
	if *dst == "" {
		*dst = fallback
	}
}
