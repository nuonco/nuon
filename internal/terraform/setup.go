package terraform

import (
	"context"
	"fmt"
	"log"
	"os"

	s3fetch "github.com/powertoolsdev/go-fetch/pkg/s3"
	"github.com/powertoolsdev/go-terraform/internal/terraform/config"
	"github.com/powertoolsdev/go-terraform/internal/terraform/config/backend"
	"github.com/powertoolsdev/go-terraform/internal/terraform/config/vars"
	"github.com/powertoolsdev/go-terraform/internal/terraform/executor"
	"github.com/powertoolsdev/go-terraform/internal/terraform/install"
	"github.com/powertoolsdev/go-terraform/internal/terraform/manager"
	"github.com/powertoolsdev/go-terraform/internal/terraform/module"
)

type tfExecutor interface {
	Init(context.Context) error
	Plan(context.Context) error
	Apply(context.Context) error
	Destroy(context.Context) error
	Output(context.Context) (map[string]interface{}, error)
}

// Setup will set up the workspace
func (w *workspace) Setup(ctx context.Context) error {
	if err := w.installTerraform(ctx); err != nil {
		return err
	}

	if err := w.setupWorkingDirectory(ctx); err != nil {
		return err
	}

	if err := w.fetchModule(ctx); err != nil {
		return err
	}

	if err := w.writeConfiguration(); err != nil {
		return err
	}

	tfExec, err := w.setupExecutor()
	if err != nil {
		return err
	}
	w.tfExecutor = tfExec

	return nil
}

func (w *workspace) installTerraform(ctx context.Context) error {
	tmpdir, err := os.MkdirTemp("", w.ID)
	if err != nil {
		return err
	}

	installer, err := install.New(
		w.validator,
		install.WithInstallDir(tmpdir),
		// TODO(jdt): fix logger
		install.WithLogger(log.Default()),
	)
	if err != nil {
		return err
	}
	w.cleanupFns = append(w.cleanupFns, installer.Cleanup)

	tfExecPath, err := installer.Install(ctx)
	if err != nil {
		return err
	}
	w.tfExecPath = tfExecPath
	return nil
}

func (w *workspace) setupWorkingDirectory(ctx context.Context) error {
	mgr, err := manager.New(w.validator, manager.WithID(w.ID))
	if err != nil {
		return err
	}
	w.workspaceWriter = mgr

	cleanup, err := mgr.Init(ctx)
	if err != nil {
		return err
	}
	w.cleanupFns = append(w.cleanupFns, cleanup)

	wd, err := mgr.GetWorkingDir()
	if err != nil {
		return err
	}
	w.workingDir = wd

	return nil
}

func (w *workspace) fetchModule(ctx context.Context) error {
	fetcher, err := s3fetch.New(
		w.validator,
		s3fetch.WithBucketName(w.Module.Bucket),
		s3fetch.WithBucketKey(w.Module.Key),
		s3fetch.WithRoleARN(w.Module.AssumeRoleDetails.AssumeArn),
		s3fetch.WithRoleSessionName(fmt.Sprintf("go-terraform-workspace-%s", w.ID)),
	)
	if err != nil {
		return err
	}

	m, err := module.New(
		w.validator,
		module.WithFetcher(fetcher),
		module.WithWriteFactory(w.workspaceWriter),
	)
	if err != nil {
		return err
	}
	return m.Install(ctx)
}

func (w *workspace) writeConfiguration() error {
	bec, err := backend.NewS3Configurator(
		w.validator,
		backend.WithBackendConfig(&backend.S3Config{
			BucketName:   w.Backend.Bucket,
			BucketKey:    w.Backend.Key,
			BucketRegion: w.Backend.Region},
		))
	if err != nil {
		return err
	}

	vc, err := vars.New(w.validator, vars.WithVars(w.Vars))
	if err != nil {
		return err
	}

	ctr, err := config.New(
		w.validator,
		config.WithWriteFactory(w.workspaceWriter),
		config.WithConfigurator(backendConfigFilename, bec),
		config.WithConfigurator(varsConfigFilename, vc),
	)
	if err != nil {
		return err
	}

	return ctr.Configure()
}

func (w *workspace) setupExecutor() (tfExecutor, error) {
	return executor.New(
		w.validator,
		executor.WithWorkingDir(w.workingDir),
		executor.WithTerraformExecPath(w.tfExecPath),
		executor.WithBackendConfigFile(backendConfigFilename),
		executor.WithVarFile(varsConfigFilename),
	)
}
