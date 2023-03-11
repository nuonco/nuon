package runner

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/terraform/internal/terraform"
	planv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/executors/v1/plan/v1"
)

// Run will actually run the terraform request by:
// - pulling the request from S3
// - parsing the request
// - setting up the workspace
// - running terraform
func (r *runner) Run(ctx context.Context) (map[string]interface{}, error) {
	defer func() { _ = r.cleanup() }()

	// setup workspace from req
	ws, err := r.workspaceSetuper.setupWorkspace(ctx, r.Plan)
	if err != nil {
		return nil, err
	}

	// execute action
	return run(ctx, ws, r.Plan.RunType)
}

type workspaceSetuper interface {
	setupWorkspace(context.Context, *planv1.TerraformPlan) (executor, error)
}

var _ workspaceSetuper = (*runner)(nil)

// setupWorkspace sets up the workspace for the given request
func (r *runner) setupWorkspace(ctx context.Context, req *planv1.TerraformPlan) (executor, error) {
	ws, err := terraform.NewWorkspace(
		r.validator,
		terraform.WithID(req.Id),
		terraform.WithModuleBucket(req.Module),
		terraform.WithBackendBucket(req.Backend),
		terraform.WithVars(req.Vars.AsMap()),
		terraform.WithVersion(req.TerraformVersion),
	)
	// NOTE(jdt): always cleanup even if error
	r.cleanupFns = append(r.cleanupFns, ws.Cleanup)
	if err != nil {
		return nil, err
	}

	if err := ws.Setup(ctx); err != nil {
		return nil, err
	}
	return ws, nil
}

type executor interface {
	Init(context.Context) error
	Apply(context.Context) error
	Plan(context.Context) error
	Destroy(context.Context) error
	Output(context.Context) (map[string]interface{}, error)
}

// run executes terraform for typ
func run(ctx context.Context, e executor, typ planv1.TerraformRunType) (map[string]interface{}, error) {
	var m map[string]interface{}

	// TODO(jdt): maybe don't init if not a valid run type?
	if err := e.Init(ctx); err != nil {
		return m, err
	}

	switch typ {
	case planv1.TerraformRunType_TERRAFORM_RUN_TYPE_PLAN:
		err := e.Plan(ctx)
		return m, err

	case planv1.TerraformRunType_TERRAFORM_RUN_TYPE_DESTROY:
		err := e.Destroy(ctx)
		return m, err

	// NOTE(jdt): there's not really a good reason to run plan
	// before apply as we essentially just auto-apply...
	case planv1.TerraformRunType_TERRAFORM_RUN_TYPE_APPLY:
		err := e.Apply(ctx)
		if err != nil {
			return m, err
		}
		return e.Output(ctx)

	default:
		return m, fmt.Errorf("invalid run type did not match any cases: %s", typ)
	}
}
