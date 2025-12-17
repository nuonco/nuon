package actions

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	temporalanalytics "github.com/nuonco/nuon/pkg/analytics/temporal"
	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	teventloop "github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

const (
	defaultNamespace string = "installs"
)

type Params struct {
	fx.In

	Cfg       *internal.Config
	DB        *gorm.DB `name:"psql"`
	V         *validator.Validate
	MW        metrics.Writer
	EVClient  teventloop.Client
	Analytics temporalanalytics.Writer
	Templates *cloudformation.Templates
}

type Workflows struct {
	cfg       *internal.Config
	v         *validator.Validate
	mw        tmetrics.Writer
	evClient  teventloop.Client
	analytics temporalanalytics.Writer
	templates *cloudformation.Templates
	db        *gorm.DB
}

func (w *Workflows) All() []any {
	return append(w.ListWorkflowFns(), w.ActionEventLoop)
}

func NewWorkflows(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MW),
		tmetrics.WithTags(map[string]string{
			"namespace":    defaultNamespace,
			"context":      "worker",
			"actor-object": "install-action-workflow",
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal metrics writer: %w", err)
	}

	return &Workflows{
		cfg:       params.Cfg,
		v:         params.V,
		evClient:  params.EVClient,
		mw:        tmw,
		analytics: params.Analytics,
		templates: params.Templates,
		db:        params.DB,
	}, nil
}
