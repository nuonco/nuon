package monitor

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	fetchtoken "github.com/nuonco/nuon/bins/runner/internal/jobs/management/fetch_token"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"

	"text/template"

	"github.com/fidiego/systemctl"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// NOTE: the process will require ownership of /opt/nuon/runner and its children
const (
	ConfigDirectory     = "/opt/nuon/runner"
	ImageConfigFilename = "/opt/nuon/runner/image"
	// systemd
	RunnerServiceDir  = "/etc/systemd/system"
	RunnerServiceName = "nuon-runner.service"
)

var defaultSystemctlOpts = systemctl.Options{UserMode: false}

//go:embed templates/image.env
var imageConfigTemplate string

//go:embed templates/runner-service.aws.service
var runnerServiceAWS string

//go:embed templates/runner-service.gcp.service
var runnerServiceGCP string

//go:embed templates/runner-service.azure.service
var runnerServiceAzure string

func (h *Monitor) checkRunnerService(ctx context.Context) error {
	h.l.Info("checking runner service")

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

	err = h.ensureRunnerTokenValid(ctx)
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
	h.l.Info(fmt.Sprintf("whoami: %s", output))
	return nil
}

func (h *Monitor) ensureConfigDirectories(ctx context.Context) error {
	h.l.Debug(fmt.Sprintf("ensuring config directory exists: %s", ConfigDirectory))
	// ensure the config dir exists: this dir may be created by the init script
	_, err := os.Stat(ConfigDirectory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			h.l.Warn("directory file does not exist - will create")
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
	h.l.Debug(fmt.Sprintf("ensuring runner image config file exists: %s", ImageConfigFilename))
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

func (h *Monitor) ensureRunnerTokenValid(ctx context.Context) error {
	h.l.Debug("ensuring runner token is valid")
	_, err := h.apiClient.GetRunner(ctx)
	if err == nil {
		return nil
	}

	if !nuonrunner.IsUnauthorized(err) && !nuonrunner.IsForbidden(err) {
		return errors.Wrap(err, "unable to validate runner token")
	}

	h.l.Warn("runner token is invalid - fetching new token via IMDS",
		zap.String("platform", h.settings.Platform))

	unauthClient, err := nuonrunner.New(
		nuonrunner.WithURL(h.settings.Cfg.RunnerAPIURL),
	)
	if err != nil {
		return errors.Wrap(err, "unable to create unauthenticated client")
	}

	var result *fetchtoken.FetchTokenResult
	switch h.settings.Platform {
	case "azure":
		result, err = fetchtoken.FetchTokenAzure(ctx, unauthClient, h.settings.Cfg.RunnerID)
	default:
		result, err = fetchtoken.FetchToken(ctx, unauthClient)
	}
	if err != nil {
		return errors.Wrap(err, "unable to fetch new token")
	}

	// Update the in-memory token on both the API client and config.
	h.apiClient.SetAuthToken(result.Token)
	h.settings.Cfg.RunnerAPIToken = result.Token

	h.l.Info(fmt.Sprintf("successfully refreshed runner token for runner %s", result.RunnerID))
	return nil
}

func (h *Monitor) ensureRunnerServiceDefinition(ctx context.Context) error {
	// NOTE(fd): we need to pivot on the runner platform to grab the right template
	path := filepath.Join(RunnerServiceDir, RunnerServiceName)
	h.l.Debug(fmt.Sprintf("ensuring runner unit file exists: %s", path))

	// dynamically choose the template based on cloud platform
	var serviceTemplate string
	switch h.settings.Platform {
	case "aws", "":
		serviceTemplate = runnerServiceAWS
	case "gcp":
		serviceTemplate = runnerServiceGCP
	case "azure":
		serviceTemplate = runnerServiceAzure
	default:
		serviceTemplate = runnerServiceAWS
	}
	tmpl := template.Must(template.New("").Parse(serviceTemplate))

	var shouldWrite bool
	// check the nuon-runner.service file
	// 1. if it does not exist, create it and write the template
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			h.l.Warn(fmt.Sprintf("the file (%s) does not exist - will create it", path))
			shouldWrite = true
		} else {
			return errors.Wrap(err, fmt.Sprintf("unable to stat %s", path))
		}
	}
	// 2. if it exists, but it is empty, overwrite it w/ the template
	if info != nil && info.Size() == 0 {
		h.l.Info(fmt.Sprintf("the file (%s) exists, but it is empty - will overwrite it", path))
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
	h.l.Debug("ensuring runner service is active")
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
		h.l.Info(fmt.Sprintf("service is up and running - uptime: %s", time))
	}

	return nil
}

func EnsureImageConfigFile(ctx context.Context, l *zap.Logger, settings *settings.Settings) error {
	// NOTE(fd): this method just writes the settings no matter what
	// TODO: we should really be comparing the settings to the contents of the file and writing only when they have changed
	l.Debug(fmt.Sprintf("ensuring runner image config file exists: %s", ImageConfigFilename))
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
