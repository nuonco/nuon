package helpers

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// RecordInput is the unified write entrypoint for the helpers.
//
// The caller supplies a typed CompositeError instance (often produced by a
// parser) plus the polymorphic owner it attaches to and an optional tree of
// causes that will be persisted in the same transaction.
type RecordInput struct {
	OwnerType string
	OwnerID   string

	// Required. The typed error instance. Must implement the
	// composite_error.CompositeError interface; its Type() must be
	// registered in the catalog.
	Error composite_error.CompositeError

	// Optional small parser-input snippet + parser identification.
	Source composite_error.Source

	// Optional pointers at other entities the UI dereferences at read time.
	References []composite_error.Reference

	// Optional causes recorded recursively. Each becomes its own row,
	// linked from the parent via composite_error_causes.
	Causes []RecordCause
}

// RecordCause is a child node in the cause tree. Mirrors RecordInput minus
// owner — children inherit the owner of the root from the helper.
type RecordCause struct {
	Error      composite_error.CompositeError
	Source     composite_error.Source
	References []composite_error.Reference
	Causes     []RecordCause
	IsPrimary  bool
}

// Record persists the error (and its causes recursively) in a single
// transaction. Returns the persisted root row.
//
// Validation:
//   - OwnerType / OwnerID must be present.
//   - Error must be non-nil and its Type() registered in the catalog.
//   - If the type implements ErrorWithJSONSchema, Data is validated.
//
// At most one IsPrimary=true cause per parent is enforced (silently coerced
// to false on extras after the first).
func (h *Helpers) Record(ctx context.Context, in RecordInput) (*app.CompositeError, error) {
	if err := validateRecordInput(in); err != nil {
		return nil, err
	}

	var root *app.CompositeError
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := h.persistOne(ctx, tx, in.OwnerType, in.OwnerID, in.Error, in.Source, in.References)
		if err != nil {
			return err
		}
		root = row

		if err := h.persistCauses(ctx, tx, root.ID, in.OwnerType, in.OwnerID, in.Causes); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return root, nil
}

// RecordFromError is the convenience entry point used by step finalizers
// (and any other failure path). It runs the parser pipeline against the raw
// inputs and records the result tree.
//
// Returns the persisted root row plus the typed error instance for callers
// that want to apply OverrideDirective() inline (e.g. the conductor).
func (h *Helpers) RecordFromError(
	ctx context.Context,
	ownerType, ownerID string,
	parseCtx composite_error.ParseContext,
	in composite_error.ParseInput,
) (*app.CompositeError, composite_error.CompositeError, error) {
	res := h.pipeline.Parse(ctx, parseCtx, in)
	rec := recordInputFromParseResult(ownerType, ownerID, res)
	row, err := h.Record(ctx, rec)
	if err != nil {
		return nil, nil, errors.Wrap(err, "record from error")
	}
	return row, res.Primary.Error, nil
}

// persistOne marshals a typed CompositeError into a row and inserts it.
// Renders Title/Summary into the cached columns at write time.
func (h *Helpers) persistOne(
	ctx context.Context,
	tx *gorm.DB,
	ownerType, ownerID string,
	ce composite_error.CompositeError,
	source composite_error.Source,
	refs []composite_error.Reference,
) (*app.CompositeError, error) {
	if err := h.validateAgainstSchema(ce); err != nil {
		return nil, err
	}

	data, err := json.Marshal(ce)
	if err != nil {
		return nil, errors.Wrap(err, "marshal composite error data")
	}

	rendered := ce.Render(ctx)

	row := &app.CompositeError{
		OwnerID:       ownerID,
		OwnerType:     ownerType,
		Type:          ce.Type(),
		Domain:        ce.Domain(),
		Severity:      ce.Severity(),
		Data:          data,
		Source:        source,
		References:    composite_error.References(refs),
		TitleCached:   rendered.Title,
		SummaryCached: rendered.Summary,
	}

	if err := tx.Create(row).Error; err != nil {
		return nil, errors.Wrap(err, "insert composite_errors row")
	}
	return row, nil
}

// persistCauses recursively persists child errors and the edges that link
// them to parentID. At most one IsPrimary=true edge survives per parent.
func (h *Helpers) persistCauses(
	ctx context.Context,
	tx *gorm.DB,
	parentID, ownerType, ownerID string,
	causes []RecordCause,
) error {
	primaryAssigned := false
	for i, c := range causes {
		if c.Error == nil {
			continue
		}
		childRow, err := h.persistOne(ctx, tx, ownerType, ownerID, c.Error, c.Source, c.References)
		if err != nil {
			return err
		}

		isPrimary := c.IsPrimary && !primaryAssigned
		if isPrimary {
			primaryAssigned = true
		}

		edge := &app.CompositeErrorCause{
			ParentID:  parentID,
			ChildID:   childRow.ID,
			Idx:       i,
			IsPrimary: isPrimary,
		}
		if err := tx.Create(edge).Error; err != nil {
			return errors.Wrap(err, "insert composite_error_causes edge")
		}

		// Recurse: each cause may itself have causes.
		if len(c.Causes) > 0 {
			if err := h.persistCauses(ctx, tx, childRow.ID, ownerType, ownerID, c.Causes); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRecordInput(in RecordInput) error {
	if in.OwnerType == "" {
		return errors.New("composite_errors.Record: OwnerType is required")
	}
	if in.OwnerID == "" {
		return errors.New("composite_errors.Record: OwnerID is required")
	}
	if in.Error == nil {
		return errors.New("composite_errors.Record: Error is required")
	}
	return nil
}

// validateAgainstSchema is a no-op when the type doesn't implement
// ErrorWithJSONSchema. When it does, validation lands here. Phase 1 keeps
// this as a placeholder — schema validation is a follow-up.
func (h *Helpers) validateAgainstSchema(ce composite_error.CompositeError) error {
	if _, ok := ce.(composite_error.ErrorWithJSONSchema); !ok {
		return nil
	}
	// TODO: wire a JSON Schema validator (e.g. xeipuuv/gojsonschema). For
	// now we trust the parser to produce well-formed Data.
	return nil
}

// recordInputFromParseResult flattens a Pipeline result into a RecordInput
// tree. The Primary becomes the root; Secondaries become non-primary causes
// at the root level alongside the Primary's own Causes.
func recordInputFromParseResult(
	ownerType, ownerID string,
	res composite_error.PipelineResult,
) RecordInput {
	in := RecordInput{
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		Error:      res.Primary.Error,
		Source:     res.Primary.Source,
		References: res.Primary.Refs,
	}

	// Causes from the primary parser's own ParseResult tree (these were the
	// root cause(s) the parser attached). Mark the first as primary if the
	// parser tagged any.
	for i, c := range res.Primary.Causes {
		in.Causes = append(in.Causes, recordCauseFromParseResult(c, i == 0))
	}

	// Cross-parser secondaries become additional non-primary causes on the
	// root, preserving the multi-error hint without flattening the tree.
	for _, s := range res.Secondaries {
		in.Causes = append(in.Causes, recordCauseFromParseResult(s, false))
	}
	return in
}

func recordCauseFromParseResult(r composite_error.ParseResult, primary bool) RecordCause {
	c := RecordCause{
		Error:      r.Error,
		Source:     r.Source,
		References: r.Refs,
		IsPrimary:  primary,
	}
	for i, sub := range r.Causes {
		c.Causes = append(c.Causes, recordCauseFromParseResult(sub, i == 0))
	}
	return c
}
