package activities

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/mapstructure"
	"go.temporal.io/sdk/activity"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/terraform/archive/dir"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
	s3backend "github.com/powertoolsdev/mono/pkg/terraform/backend/s3"
	remotebinary "github.com/powertoolsdev/mono/pkg/terraform/binary/remote"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks/noop"
	"github.com/powertoolsdev/mono/pkg/terraform/outputs"
	"github.com/powertoolsdev/mono/pkg/terraform/run"
	staticvars "github.com/powertoolsdev/mono/pkg/terraform/variables/static"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

const (
	defaultTerraformVersion  string        = "v1.8.4"
	defaultHeartBeatInterval time.Duration = time.Second * 1
)

type RunType string

const (
	RunTypeDestroy RunType = "destroy"
	RunTypeApply   RunType = "apply"
)

type TerraformRunOutputs struct {
	OrgID string `mapstructure:"org_id" json:"org_id"`

	AppID string                 `mapstructure:"app_id" json:"app_id"`
	App   map[string]interface{} `mapstructure:"app" json:"app"`

	ComponentIDs []string               `mapstructure:"component_ids" json:"component_ids"`
	Components   map[string]interface{} `mapstructure:"components" json:"components"`

	InstallIDs []string      `mapstructure:"install_ids" json:"install_ids"`
	Installs   []interface{} `mapstructure:"installs" json:"installs"`
}

type RunTerraformRequest struct {
	RunType      RunType `validate:"required"`
	CanaryID     string  `validate:"required"`
	OrgID        string  `validate:"required"`
	APIToken     string  `validate:"required"`
	InstallCount int     `validate:"required"`
}

type RunTerraformResponse struct {
	Outputs *TerraformRunOutputs
}

func (a *Activities) getWorkspace(moduleDir string, req *RunTerraformRequest) (workspace.Workspace, error) {
	arch, err := dir.New(a.v,
		dir.WithPath(moduleDir),
		dir.WithIgnoreDotTerraformDir(),
		dir.WithIgnoreTerraformStateFile(),
		dir.WithIgnoreTerraformLockFile(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create archive: %w", err)
	}

	bin, err := remotebinary.New(a.v,
		remotebinary.WithVersion(defaultTerraformVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	vars, err := staticvars.New(a.v, staticvars.WithFileVars(map[string]interface{}{
		"aws_eks_iam_role_arn":      a.cfg.AWSEKSIAMRoleArn,
		"aws_ecs_iam_role_arn":      a.cfg.AWSECSIAMRoleArn,
		"azure_aks_subscription_id": a.cfg.AzureAKSSubscriptionID,
		"azure_aks_tenant_id":       a.cfg.AzureAKSTenantID,
		"azure_aks_client_id":       a.cfg.AzureAKSClientID,
		"azure_aks_client_secret":   a.cfg.AzureAKSClientSecret,
		"install_count":             req.InstallCount,
	}),
		staticvars.WithEnvVars(map[string]string{
			"NUON_ORG_ID":    req.OrgID,
			"NUON_API_URL":   a.cfg.APIURL,
			"NUON_API_TOKEN": req.APIToken,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create vars: %w", err)
	}

	var back backend.Backend
	back, err = s3backend.New(a.v,
		s3backend.WithCredentials(&credentials.Config{
			UseDefault: true,
		}),
		s3backend.WithBucketConfig(&s3backend.BucketConfig{
			Name:   a.cfg.StateBucketName,
			Key:    fmt.Sprintf("%s/state.tfstate", req.CanaryID),
			Region: a.cfg.StateBucketRegion,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create backend: %w", err)
	}

	hooks := noop.New()

	// create workspace
	wkspace, err := workspace.New(a.v,
		workspace.WithHooks(hooks),
		workspace.WithArchive(arch),
		workspace.WithBackend(back),
		workspace.WithBinary(bin),
		workspace.WithVariables(vars),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}
	return wkspace, nil
}

func (a *Activities) runTerraform(ctx context.Context, moduleDir string, req *RunTerraformRequest) (map[string]interface{}, error) {
	wkspace, err := a.getWorkspace(moduleDir, req)
	if err != nil {
		return nil, fmt.Errorf("unable to get workspace: %w", err)
	}

	runLog := hclog.New(&hclog.LoggerOptions{
		Name:   "terraform",
		Output: os.Stderr,
	})

	tfRun, err := run.New(a.v,
		run.WithWorkspace(wkspace),
		run.WithLogger(runLog),
		run.WithOutputSettings(&run.OutputSettings{
			Ignore: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create run: %w", err)
	}

	if req.RunType == RunTypeDestroy {
		err = tfRun.Destroy(ctx)
	} else {
		err = tfRun.Apply(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to run: %w", err)
	}

	wkspaceOutputs, err := wkspace.Output(ctx, runLog)
	if err != nil {
		return nil, fmt.Errorf("unable to get outputs: %w", err)
	}
	out, err := outputs.TFOutputMetaToStructPB(wkspaceOutputs)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to standard outputs: %w", err)
	}

	return out.AsMap(), nil
}

func (a *Activities) RunTerraform(ctx context.Context, req *RunTerraformRequest) (*RunTerraformResponse, error) {
	// TODO(jm): this heartbeating step should probably be pulled into pkg/workflows
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()

	ch := make(chan struct{})
	defer close(ch)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(defaultHeartBeatInterval)
		for {
			select {
			case <-ch:
				ticker.Stop()
				return
			case <-ticker.C:
				activity.RecordHeartbeat(ctx, map[string]interface{}{})
			}
		}
	}()

	moduleDir := a.cfg.TerraformModuleDir
	rawOutputs, err := a.runTerraform(ctx, moduleDir, req)
	if err != nil {
		return nil, fmt.Errorf("unable to run terraform: %w", err)
	}

	var outputs TerraformRunOutputs
	if err := mapstructure.Decode(rawOutputs, &outputs); err != nil {
		return nil, fmt.Errorf("unable to parse outputs: %w", err)
	}
	outputs.OrgID = req.OrgID

	return &RunTerraformResponse{
		Outputs: &outputs,
	}, nil
}
