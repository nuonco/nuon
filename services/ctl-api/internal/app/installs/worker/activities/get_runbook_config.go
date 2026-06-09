package activities

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetRunbookConfigByIDRequest struct {
	RunbookConfigID   string `validate:"required"`
	InstallWorkflowID string
}

// @temporal-gen-v2 activity
// @by-field RunbookConfigID
func (a *Activities) GetRunbookConfigByID(ctx context.Context, req GetRunbookConfigByIDRequest) (*app.RunbookConfig, error) {
	var rbConfig app.RunbookConfig
	res := a.db.WithContext(ctx).
		Preload("Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Inputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		First(&rbConfig, "id = ?", req.RunbookConfigID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runbook config: %w", res.Error)
	}

	if req.InstallWorkflowID == "" {
		return &rbConfig, nil
	}

	var run app.InstallRunbookRun
	if err := a.db.WithContext(ctx).
		First(&run, "install_workflow_id = ?", req.InstallWorkflowID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install runbook run: %w", err)
	}

	inputs := map[string]string{}
	for k, v := range run.RunbookInputs {
		if v != nil {
			inputs[k] = *v
		}
	}

	data := map[string]any{"runbook_inputs": inputs}
	for i := range rbConfig.Steps {
		if err := renderRunbookStep(&rbConfig.Steps[i], data); err != nil {
			return nil, fmt.Errorf("unable to render runbook step %s: %w", rbConfig.Steps[i].Name, err)
		}
	}

	return &rbConfig, nil
}

func renderRunbookStep(step *app.RunbookStepConfig, data map[string]any) error {
	var err error
	if step.Command, err = renderRunbookInput(step.Command, data); err != nil {
		return err
	}
	if step.InlineContents, err = renderRunbookInput(step.InlineContents, data); err != nil {
		return err
	}
	if step.Role, err = renderRunbookInput(step.Role, data); err != nil {
		return err
	}
	for k, v := range step.EnvVars {
		if v == nil {
			continue
		}
		rendered, rerr := renderRunbookInput(*v, data)
		if rerr != nil {
			return rerr
		}
		step.EnvVars[k] = &rendered
	}
	return nil
}

// renderRunbookInput renders a single field against the runbook_inputs data map.
// Fields without a runbook_inputs reference are returned unchanged so that any
// .nuon templating is left intact for later execution stages.
func renderRunbookInput(field string, data map[string]any) (string, error) {
	if !strings.Contains(field, "runbook_inputs") {
		return field, nil
	}

	tmpl, err := template.New("runbook_input").
		Funcs(sprig.TxtFuncMap()).
		Option("missingkey=error").
		Parse(field)
	if err != nil {
		return field, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return field, err
	}

	return buf.String(), nil
}
