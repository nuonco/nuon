package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	s3fetch "github.com/powertoolsdev/go-fetch/pkg/s3"
	"github.com/powertoolsdev/go-terraform/internal/terraform"
)

// Run will actually run the terraform request by:
// - pulling the request from S3
// - parsing the request
// - setting up the workspace
// - running terraform
func (r *runner) Run(ctx context.Context) (map[string]interface{}, error) {
	defer func() { _ = r.cleanup() }()

	// get S3 reader
	iorc, err := r.moduleFetcher.fetchModule(ctx)
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
	return run(ctx, ws, req.RunType)
}

type moduleFetcher interface {
	fetchModule(context.Context) (io.ReadCloser, error)
}

var _ moduleFetcher = (*runner)(nil)

// fetchModule pulls the module from S3
func (r *runner) fetchModule(ctx context.Context) (io.ReadCloser, error) {
	f, err := s3fetch.New(
		r.validator,
		s3fetch.WithBucketName(r.Bucket),
		s3fetch.WithBucketKey(r.Key),
		// s3.WithRegion(r.Region),
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

// Request is the necessary information for running terraform
type Request struct {
	// ID is the opaque identifier for the run.
	// Historically, this was the installation short ID
	ID string `json:"id"`
	// Sandbox is the cloud object for the sandbox module to use (tar and gzipped)
	Sandbox Object `json:"sandbox"`
	// Backend is the cloud object for the backend state store
	Backend Object `json:"backend"`
	// Vars are the terraform vars
	Vars map[string]interface{} `json:"vars"`
	// RunType is the type of run being requested
	RunType RunType `json:"run_type"`
}

type requestParser interface {
	parseRequest(io.Reader) (*Request, error)
}

var _ requestParser = (*runner)(nil)

// parseRequest parses the request
// typically, this would be pulled from S3
func (r *runner) parseRequest(ior io.Reader) (*Request, error) {
	bs, err := io.ReadAll(ior)
	if err != nil {
		return nil, err
	}

	var req Request
	err = json.Unmarshal(bs, &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

type workspaceSetuper interface {
	setupWorkspace(context.Context, *Request) (executor, error)
}

var _ workspaceSetuper = (*runner)(nil)

// setupWorkspace sets up the workspace for the given request
func (r *runner) setupWorkspace(ctx context.Context, req *Request) (executor, error) {
	sb := terraform.Object(req.Sandbox)
	bb := terraform.Object(req.Backend)

	ws, err := terraform.NewWorkspace(
		r.validator,
		terraform.WithID(req.ID),
		terraform.WithSandboxBucket(&sb),
		terraform.WithBackendBucket(&bb),
		terraform.WithVars(req.Vars),
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
