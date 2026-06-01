package stack

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	temporalanalytics "github.com/nuonco/nuon/pkg/analytics/temporal"
	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/arm"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

const (
	defaultNamespace string = "installs"
)

type Params struct {
	fx.In

	Cfg          *internal.Config
	DB           *gorm.DB `name:"psql"`
	V            *validator.Validate
	MW           metrics.Writer
	Analytics    temporalanalytics.Writer
	Templates    *cloudformation.Templates
	ARMTemplates *arm.Templates
}

type Workflows struct {
	cfg          *internal.Config
	v            *validator.Validate
	mw           tmetrics.Writer
	analytics    temporalanalytics.Writer
	templates    *cloudformation.Templates
	armTemplates *arm.Templates
	db           *gorm.DB
}

func (w *Workflows) All() []any {
	return []any{
		w.GenerateInstallStackVersion,
	}
}

func NewWorkflows(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MW),
		tmetrics.WithTags(map[string]string{
			"namespace":    defaultNamespace,
			"context":      "worker",
			"actor-object": "install-stack",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}

	return &Workflows{
		cfg:          params.Cfg,
		v:            params.V,
		mw:           tmw,
		analytics:    params.Analytics,
		templates:    params.Templates,
		armTemplates: params.ARMTemplates,
		db:           params.DB,
	}, nil
}
