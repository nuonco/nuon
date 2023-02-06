package runner

import (
	"context"
	"fmt"
	"io"

	s3fetch "github.com/powertoolsdev/go-fetch/pkg/s3"
	"github.com/powertoolsdev/go-terraform/internal/terraform"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"google.golang.org/protobuf/proto"
)

// Run will actually run the terraform request by:
// - pulling the request from S3
// - parsing the request
// - setting up the workspace
// - running terraform
func (r *runner) Run(ctx context.Context) (map[string]interface{}, error) {
	defer func() { _ = r.cleanup() }()

	// get S3 reader
	iorc, err := r.planFetcher.fetchPlan(ctx)
	if err != nil {
		return nil, err
	}
	r.cleanupFns = append(r.cleanupFns, iorc.Close)

	// read and parse
	req, err := r.requestParser.parseRequest(iorc)
	if err != nil {
		return nil, err
	}

	// setup workspace from req
	ws, err := r.workspaceSetuper.setupWorkspace(ctx, req)
	if err != nil {
		return nil, err
	}

	// execute action
	return run(ctx, ws, RunTypeApply)
}

type planFetcher interface {
	fetchPlan(context.Context) (io.ReadCloser, error)
}

var _ planFetcher = (*runner)(nil)

// fetchPlan pulls the plan from S3
func (r *runner) fetchPlan(ctx context.Context) (io.ReadCloser, error) {
	f, err := s3fetch.New(
		r.validator,
		s3fetch.WithBucketName(r.Plan.Bucket),
		s3fetch.WithBucketKey(r.Plan.BucketKey),
		s3fetch.WithRoleARN(r.Plan.BucketAssumeRoleArn),
	)
	if err != nil {
		return nil, err
	}
	return f.Fetch(ctx)
}

// TODO(jdt): move to different module / repo

// RunType is the type of run being requested
// Corresponds to the equivalent terraform commands
type RunType string

const (
	RunTypeApply   RunType = "apply"
	RunTypePlan    RunType = "plan"
	RunTypeDestroy RunType = "destroy"
)

// Object represents an object in cloud storage (e.g. S3)
type Object struct {
	BucketName string `json:"bucket"`
	Key        string `json:"key"`
	Region     string `json:"region"`
}

type requestParser interface {
	parseRequest(io.Reader) (*planv1.TerraformPlan, error)
}

var _ requestParser = (*runner)(nil)

// parseRequest parses the request
// typically, this would be pulled from S3
func (r *runner) parseRequest(ior io.Reader) (*planv1.TerraformPlan, error) {
	bs, err := io.ReadAll(ior)
	if err != nil {
		return nil, err
	}

	var tfp planv1.TerraformPlan
	if err = proto.Unmarshal(bs, &tfp); err != nil {
		return nil, err
	}
	return &tfp, nil
}

type workspaceSetuper interface {
	setupWorkspace(context.Context, *planv1.TerraformPlan) (executor, error)
}

var _ workspaceSetuper = (*runner)(nil)

// setupWorkspace sets up the workspace for the given request
func (r *runner) setupWorkspace(ctx context.Context, req *planv1.TerraformPlan) (executor, error) {
	vars := map[string]interface{}{}
	for k, v := range req.Vars {
		vars[k] = v
	}

	ws, err := terraform.NewWorkspace(
		r.validator,
		terraform.WithID(req.Id),
		terraform.WithModuleBucket(req.Module),
		terraform.WithBackendBucket(req.Backend),
		terraform.WithVars(vars),
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
func run(ctx context.Context, e executor, typ RunType) (map[string]interface{}, error) {
	var m map[string]interface{}

	// TODO(jdt): maybe don't init if not a valid run type?
	if err := e.Init(ctx); err != nil {
		return m, err
	}

	switch typ {
	case RunTypePlan:
		err := e.Plan(ctx)
		return m, err

	case RunTypeDestroy:
		err := e.Destroy(ctx)
		return m, err

	// NOTE(jdt): there's not really a good reason to run plan
	// before apply as we essentially just auto-apply...
	case RunTypeApply:
		err := e.Apply(ctx)
		if err != nil {
			return m, err
		}
		return e.Output(ctx)

	default:
		return m, fmt.Errorf("invalid run type did not match any cases: %s", typ)
	}
}
