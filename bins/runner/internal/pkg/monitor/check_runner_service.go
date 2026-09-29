package monitor

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	fetchtoken "github.com/nuonco/nuon/bins/runner/internal/jobs/management/fetch_token"
	"github.com/nuonco/nuon/pkg/runner/settings"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	ConfigDirectory                 = "/opt/nuon/runner"
	ImageConfigFilename             = "/opt/nuon/runner/image"
	TelemetryExportStorageDirectory = "/opt/nuon/runner/telemetry-export"
	telemetryExportStorageMount     = TelemetryExportStorageDirectory + ":/var/lib/nuon/telemetry-export"
)

//go:embed templates/image.env
var imageConfigTemplate string

func (h *Monitor) checkRunnerService(ctx context.Context) error {
	h.l.Info("checking runner service")

	err := h.whoami(ctx)
	if err != nil {
		h.l.Error(err.Error())
		return err
	}

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

	err = h.ensureRunnerSystemdService(ctx)
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
		result, err = fetchtoken.FetchToken(ctx, unauthClient, h.settings.Cfg.RunnerAuthMethod, h.settings.Cfg.RunnerID)
	}
	if err != nil {
		return errors.Wrap(err, "unable to fetch new token")
	}

	h.apiClient.SetAuthToken(result.Token)
	h.settings.Cfg.RunnerAPIToken = result.Token

	h.l.Info(fmt.Sprintf("successfully refreshed runner token for runner %s", result.RunnerID))
	return nil
}

func EnsureImageConfigFile(ctx context.Context, l *zap.Logger, settings *settings.Settings) error {
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
