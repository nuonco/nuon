package components

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	temporalanalytics "github.com/powertoolsdev/mono/pkg/analytics/temporal"
	"github.com/powertoolsdev/mono/pkg/metrics"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/actions"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cloudformation"
	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
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
	// FIXME(sdboyer) remove ASAP, once lifecycle workflows are deprecated
	ActionsWorkflows *actions.Workflows
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
	return append(w.ListWorkflowFns(),
		w.ComponentEventLoop,
		w.DriftCheck,
	)
}

func NewWorkflows(params Params) (*Workflows, error) {
	tmw, err := tmetrics.New(params.V,
		tmetrics.WithMetricsWriter(params.MW),
		tmetrics.WithTags(map[string]string{
			"namespace":    defaultNamespace,
			"context":      "worker",
			"actor-object": "install-component",
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
