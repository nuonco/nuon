package workspace

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/go-hclog"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=client_mock_test.go -source=client.go -package=workspace

func (w *workspace) getClient(ctx context.Context, log hclog.Logger) (Terraform, error) {
	tf, err := tfexec.NewTerraform(w.root, w.execPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get terraform client: %w", err)
	}

	if err := tf.SetEnv(w.envVars); err != nil {
		return nil, fmt.Errorf("unable to set environment variables: %w", err)
	}

	tf.SetLogger(log.StandardLogger(nil))
	return tf, nil
}

type Terraform interface {
	SetEnv(env map[string]string) error
	SetStdout(w io.Writer)
	SetStderr(w io.Writer)
	SetLog(log string) error
	SetLogCore(logCore string) error
	SetLogPath(path string) error
	SetLogProvider(logProvider string) error
	SetAppendUserAgent(ua string) error
	SetDisablePluginTLS(disabled bool) error
	SetSkipProviderVerify(skip bool) error
	WorkingDir() string
	ExecPath() string
	Init(ctx context.Context, opts ...tfexec.InitOption) error
	Version(ctx context.Context, skipCache bool) (*goversion.Version, map[string]*goversion.Version, error)
	Apply(ctx context.Context, opts ...tfexec.ApplyOption) error
	ApplyJSON(ctx context.Context, w io.Writer, opts ...tfexec.ApplyOption) error
	Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error
	DestroyJSON(ctx context.Context, w io.Writer, opts ...tfexec.DestroyOption) error
	FormatString(ctx context.Context, content string) (string, error)
	Format(ctx context.Context, unformatted io.Reader, formatted io.Writer) error
	FormatWrite(ctx context.Context, opts ...tfexec.FormatOption) error
	FormatCheck(ctx context.Context, opts ...tfexec.FormatOption) (bool, []string, error)
	ForceUnlock(ctx context.Context, lockID string, opts ...tfexec.ForceUnlockOption) error
	Graph(ctx context.Context, opts ...tfexec.GraphOption) (string, error)
	Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error)
	Validate(ctx context.Context) (*tfjson.ValidateOutput, error)
	Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error)
	PlanJSON(ctx context.Context, w io.Writer, opts ...tfexec.PlanOption) (bool, error)
	ProvidersLock(ctx context.Context, opts ...tfexec.ProvidersLockOption) error
	Get(ctx context.Context, opts ...tfexec.GetCmdOption) error
	ProvidersSchema(ctx context.Context) (*tfjson.ProviderSchemas, error)
	Refresh(ctx context.Context, opts ...tfexec.RefreshCmdOption) error
	RefreshJSON(ctx context.Context, w io.Writer, opts ...tfexec.RefreshCmdOption) error
	Show(ctx context.Context, opts ...tfexec.ShowOption) (*tfjson.State, error)
	ShowStateFile(ctx context.Context, statePath string, opts ...tfexec.ShowOption) (*tfjson.State, error)
	ShowPlanFile(ctx context.Context, planPath string, opts ...tfexec.ShowOption) (*tfjson.Plan, error)
	ShowPlanFileRaw(ctx context.Context, planPath string, opts ...tfexec.ShowOption) (string, error)
	StatePull(ctx context.Context, opts ...tfexec.StatePullOption) (string, error)
	StateMv(ctx context.Context, source, destination string, opts ...tfexec.StateMvCmdOption) error
}
