package monitor

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"

	"text/template"

	"github.com/fidiego/systemctl"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// NOTE: the process will require ownership of /opt/nuon/runner and its children
const (
	ConfigDirectory     = "/opt/nuon/runner"
	ImageConfigFilename = "/opt/nuon/runner/image"
	RunnerTokenFilename = "/opt/nuon/runner/token"
	// systemd
	RunnerServiceDir  = "/etc/systemd/system"
	RunnerServiceName = "nuon-runner.service"
)

var defaultSystemctlOpts = systemctl.Options{UserMode: false}

//go:embed templates/image.env
var imageConfigTemplate string

//go:embed templates/token.env
var runnerTokenTemplate string

//go:embed templates/runner-service.aws.service
var runnerServiceAWS string

func (h *Monitor) checkRunnerService(ctx context.Context) error {
	fmt.Println("checking runner service")

	// sanity check/debug
	err := h.whoami(ctx)
	if err != nil {
		h.l.Error(err.Error())
		return err
	}

	// the basics
	err = h.ensureConfigDirectories(ctx)
	if err != nil {
		h.l.Error(err.Error())
		return err
	}

	err = h.ensureImageConfigFile(ctx)
	if err != nil {
		h.l.Error(err.Error())
		return err
	}

	err = h.ensureRunnerTokenFile(ctx)
	if err != nil {
		h.l.Error(err.Error())
		return err
	}

	// systemd stuff
	err = h.ensureRunnerServiceDefinition(ctx)
	if err != nil {
		h.l.Error(err.Error())
		return err
	}
	err = h.ensureRunnerServiceIsActive(ctx)
	if err != nil {
		h.l.Error(err.Error())
		return err
	}
	return nil
}

func (h *Monitor) whoami(ctx context.Context) error {
	cmd, err := exec.Command("whoami").Output()
	if err != nil {
		return err
	}
	output := string(cmd)
	fmt.Printf("whoami: %s\n", output)
	return nil
}

func (h *Monitor) ensureConfigDirectories(ctx context.Context) error {
	fmt.Println(fmt.Sprintf("ensuring config directory exists: %s\n", ConfigDirectory))
	// ensure the config dir exists: this dir may be created by the init script
	_, err := os.Stat(ConfigDirectory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("directory file does not exist - will create")
			err = os.Mkdir(ConfigDirectory, os.ModeDir)
			if err != nil {
				return errors.Wrap(err, "unable to create config directory")
			}
		} else {
			return errors.Wrap(err, "unable to find config directory")
		}
	}
	return nil
}

func (h *Monitor) ensureImageConfigFile(ctx context.Context) error {
	// NOTE(fd): this method just writes the settings no matter what
	// TODO: we should really be comparing the settings to the contents of the file and writing only when they have changed
	fmt.Println(fmt.Sprintf("ensuring runner image config file exists: %s\n", ImageConfigFilename))
	tmpl := template.Must(template.New("").Parse(imageConfigTemplate))
	f, err := os.Create(ImageConfigFilename)
	if err != nil {
		return errors.Wrap(err, "unable to create image config file")
	}
	err = tmpl.Execute(f, h.settings)
	if err != nil {
		return errors.Wrap(err, "unable to execute template for image config file")
	}
	f.Close()
	return nil
}

func (h *Monitor) ensureRunnerTokenFile(ctx context.Context) error {
	// NOTE(fd): this config is special - this method only checks to see if it exists,
	// it doesn't try to create it. that's for two reasons: 1) the token this runner
	// process has in-hand is expected to already be set in the token file and 2) the
	// file should only ever be over-written in case of token refresh and we haven't
	// written that code yet.
	fmt.Println(fmt.Sprintf("ensuring runner token file exists: %s\n", RunnerTokenFilename))
	_, err := os.Stat(RunnerTokenFilename)
	if err != nil {
		return errors.Wrap(err, "unable to stat runner token file")
	}
	return nil
}

func (h *Monitor) ensureRunnerServiceDefinition(ctx context.Context) error {
	// NOTE(fd): we need to pivot on the runner platform to grab the right template
	path := filepath.Join(RunnerServiceDir, RunnerServiceName)
	fmt.Println(fmt.Sprintf("ensuring runner unit file exists: %s\n", path))

	// dynamically choose the template
	var tmpl *template.Template
	if h.settings.Platform == "aws" {
		tmpl = template.Must(template.New("").Parse(runnerServiceAWS))
	}

	var shouldWrite bool
	// check the nuon-runner.service file
	// 1. if it does not exist, create it and write the template
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println(fmt.Sprintf("the file (%s) does not exist - will create it", path))
			shouldWrite = true
		} else {
			return errors.Wrap(err, fmt.Sprintf("unable to stat %s", path))
		}
	}
	// 2. if it exists, but it is empty, overwrite it w/ the template
	if info.Size() == 0 {
		fmt.Println(fmt.Sprintf("the file (%s) exists, but it is empty - will overwrite it", path))
		shouldWrite = true
	}

	// 3. otherwise, do nothing (consider checking for drift?)
	// NOTE(fd): we really only ever write the noun-runner.service file once. we never need to
	// overwrite it since the service definition is baked into the runner binary and doesn't change.
	// if we need to modify the service definition we need to release new version of the runner and
	// run instance refresh or equivalent.

	if shouldWrite {
		err := h.writeNuonRunnerService(tmpl, path)
		if err != nil {
			return err
		}
		err = systemctl.DaemonReload(ctx, systemctl.Options{})
		if err != nil {
			return errors.Wrap(err, "unabel to reload the sytemctl daemon")
		}
	} else {
		h.l.Debug(fmt.Sprintf("%s - everything is in order", RunnerServiceName))
	}

	return nil
}

func (h *Monitor) writeNuonRunnerService(tmpl *template.Template, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to create %s", path))
	}
	err = tmpl.Execute(f, h.settings)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to execute template for %s", path))
	}
	f.Close()
	return nil
}

// this method encapsulates all of the logic to ensure the service is running.
// NOTE: we use start instead of enable
func (h *Monitor) ensureRunnerServiceIsActive(ctx context.Context) error {
	fmt.Println("ensuring runner service is active")
	isActive, err := systemctl.IsActive(ctx, RunnerServiceName, defaultSystemctlOpts)
	if err != nil {
		return errors.Wrap(err, "unable to determine if unit is active")
	}
	if !isActive {
		err = systemctl.Start(ctx, RunnerServiceName, defaultSystemctlOpts)
		if err != nil {
			return errors.Wrap(err, "unable to start unit")
		}
	} else {
		time, err := systemctl.GetStartTime(ctx, RunnerServiceName, defaultSystemctlOpts)
		if err != nil {
			return errors.Wrap(err, "unable to determine start time")
		}
		fmt.Println(fmt.Sprintf("service is up and running - uptime: %s\n", time))
	}

	return nil
}

func EnsureImageConfigFile(ctx context.Context, l *zap.Logger, settings *settings.Settings) error {
	// NOTE(fd): this method just writes the settings no matter what
	// TODO: we should really be comparing the settings to the contents of the file and writing only when they have changed
	l.Info(fmt.Sprintf("ensuring runner image config file exists: %s\n", ImageConfigFilename))
	tmpl := template.Must(template.New("").Parse(imageConfigTemplate))
	f, err := os.Create(ImageConfigFilename)
	if err != nil {
		return errors.Wrap(err, "unable to create image config file")
	}
	err = tmpl.Execute(f, settings)
	if err != nil {
		return errors.Wrap(err, "unable to execute template for image config file")
	}
	f.Close()
	return nil
}
