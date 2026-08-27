package helpers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// SetInstallInputsFromStack persists the input values a stack reported as the
// install's current inputs. Only customer-source inputs may be set.
//
// Appends a row rather than mutating, and only when values differ. On a change it
// creates an input-update workflow the caller must enqueue.
func (h *Helpers) SetInstallInputsFromStack(ctx context.Context, install *app.Install, submitted map[string]string) (*app.InstallInputs, *app.Workflow, error) {
	if len(submitted) == 0 {
		return nil, nil, nil
	}

	// Pinned to the install's app config, matching the inputs POST/PATCH paths: the
	// app's newest input config may belong to a config this install is not on.
	inputCfg, err := h.GetPinnedAppInputConfig(ctx, install.AppID, install.AppConfigID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get pinned app input config: %w", err)
	}
	if inputCfg == nil || inputCfg.ID == "" {
		return nil, nil, stderr.ErrUser{
			Err:         fmt.Errorf("no app input config on app config %s", install.AppConfigID),
			Description: "no app input configs defined",
		}
	}

	customerInputs := map[string]bool{}
	for _, in := range inputCfg.AppInputs {
		if in.Source == app.AppInputSourceCustomer {
			customerInputs[in.Name] = true
		}
	}

	var unknown []string
	for name := range submitted {
		if !customerInputs[name] {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return nil, nil, stderr.ErrUser{
			Err:         fmt.Errorf("inputs not declared as customer-source app inputs: %s", strings.Join(unknown, ", ")),
			Description: "inputs are not declared as customer-source app inputs on this app: " + strings.Join(unknown, ", "),
		}
	}

	var inputs *app.InstallInputs
	var changed *ChangedInputsResult
	// read-modify-append, so serialized against the other inputs writers
	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := LockInstallInputs(ctx, tx, install.ID); err != nil {
			return err
		}

		// newest row for the install: the one readers resolve as current
		var latest app.InstallInputs
		if res := tx.WithContext(ctx).
			Where(app.InstallInputs{InstallID: install.ID}).
			Order("created_at DESC").
			Limit(1).
			Find(&latest); res.Error != nil {
			return fmt.Errorf("unable to get install inputs: %w", res.Error)
		}

		merged := map[string]*string{}
		for k, v := range latest.Values {
			merged[k] = v
		}
		submittedPtr := map[string]*string{}
		for k, v := range submitted {
			submittedPtr[k] = generics.ToPtr(v)
			merged[k] = generics.ToPtr(v)
		}
		var err error
		changed, err = ComputeChangedInputs(latest.Values, submittedPtr, inputCfg.AppInputs)
		if err != nil {
			return fmt.Errorf("unable to compute changed inputs: %w", err)
		}
		if len(changed.Names) == 0 {
			changed = nil
			return nil
		}

		inputs = &app.InstallInputs{
			AppInputConfigID: inputCfg.ID,
			InstallID:        install.ID,
			Values:           pgtype.Hstore(merged),
		}
		if err := tx.WithContext(ctx).Create(inputs).Error; err != nil {
			return fmt.Errorf("unable to create install inputs: %w", err)
		}
		// stale_at alone is inert: the partial has to be named or state serves the
		// old inputs
		return h.MarkInstallStatePartialsStale(ctx, tx, install.ID, pkgstate.PartialInputs)
	}); err != nil {
		return nil, nil, err
	}
	if changed == nil {
		return nil, nil, nil
	}

	// Same shape as the stack-outputs input path: deploy dependents, full update.
	workflow, err := h.CreateAndStartInputUpdateWorkflow(
		ctx,
		install.ID,
		changed.Names,
		changed.ChangedValuesJSON,
		"",
		true,
		false,
		false,
		app.WorkflowTypeInputUpdate,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create input update workflow: %w", err)
	}

	return inputs, workflow, nil
}
