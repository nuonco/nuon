package customermanagedmng

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/fidiego/systemctl"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/monitor"
	"go.uber.org/zap"
)

const (
	serviceName = "nuon-runner.service"
	servicePath = "/etc/systemd/system/nuon-runner.service"
)

type Options struct {
	EnvFile   string
	ImageFile string
}

type serviceData struct {
	EnvFile, ImageFile, RunnerID, AWSRegion, BundleURI, StateURI, StackOutputsURI, InstallInputsURI, Workdir, DeploymentID string
}

func Run(ctx context.Context, options Options) error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync()

	env, err := readEnvFile(options.EnvFile)
	if err != nil {
		return fmt.Errorf("read runner env file: %w", err)
	}
	image, err := readEnvFile(options.ImageFile)
	if err != nil {
		return fmt.Errorf("read runner image file: %w", err)
	}
	for _, key := range []string{"RUNNER_ID", "AWS_REGION", "CUSTOMER_MANAGED_BUNDLE_URI"} {
		if env[key] == "" {
			return fmt.Errorf("%s is required in %s", key, options.EnvFile)
		}
	}
	for _, key := range []string{"CONTAINER_IMAGE_URL", "CONTAINER_IMAGE_TAG"} {
		if image[key] == "" {
			return fmt.Errorf("%s is required in %s", key, options.ImageFile)
		}
	}
	workdir := env["CUSTOMER_MANAGED_WORKDIR"]
	if workdir == "" {
		workdir = "/tmp/customer_managed"
	}
	data := serviceData{EnvFile: options.EnvFile, ImageFile: options.ImageFile, RunnerID: env["RUNNER_ID"], AWSRegion: env["AWS_REGION"], BundleURI: env["CUSTOMER_MANAGED_BUNDLE_URI"], StateURI: env["CUSTOMER_MANAGED_STATE_URI"], StackOutputsURI: env["CUSTOMER_MANAGED_STACK_OUTPUTS_URI"], InstallInputsURI: env["CUSTOMER_MANAGED_INSTALL_INPUTS_URI"], Workdir: workdir, DeploymentID: env["CUSTOMER_MANAGED_DEPLOYMENT_ID"]}
	if err := writeService(data); err != nil {
		return err
	}
	if registry := env["CUSTOMER_MANAGED_ECR_REGISTRY"]; registry != "" {
		if err := ecrLogin(ctx, env["AWS_REGION"], registry); err != nil {
			logger.Warn("ECR login failed", zap.Error(err))
		}
	}
	if err := ensureService(ctx); err != nil {
		return err
	}

	checkTicker := time.NewTicker(30 * time.Second)
	loginTicker := time.NewTicker(6 * time.Hour)
	defer checkTicker.Stop()
	defer loginTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-checkTicker.C:
			if err := ensureService(ctx); err != nil {
				logger.Error("ensure runner service", zap.Error(err))
			}
		case <-loginTicker.C:
			if registry := env["CUSTOMER_MANAGED_ECR_REGISTRY"]; registry != "" {
				if err := ecrLogin(ctx, env["AWS_REGION"], registry); err != nil {
					logger.Warn("ECR login failed", zap.Error(err))
				}
			}
		}
	}
}

func readEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("invalid environment line %q", line)
		}
		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return values, scanner.Err()
}

func writeService(data serviceData) error {
	if err := os.MkdirAll(filepath.Dir(servicePath), 0o755); err != nil {
		return fmt.Errorf("create systemd directory: %w", err)
	}
	tmpl, err := template.New("service").Parse(monitor.RunnerServiceCustomerManagedAWS)
	if err != nil {
		return fmt.Errorf("parse systemd template: %w", err)
	}
	f, err := os.Create(servicePath)
	if err != nil {
		return fmt.Errorf("create systemd unit: %w", err)
	}
	executeErr := tmpl.Execute(f, data)
	closeErr := f.Close()
	if executeErr != nil {
		return fmt.Errorf("render systemd unit: %w", executeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close systemd unit: %w", closeErr)
	}
	return nil
}

func ensureService(ctx context.Context) error {
	opts := systemctl.Options{UserMode: false}
	if err := systemctl.DaemonReload(ctx, opts); err != nil {
		return fmt.Errorf("reload systemd: %w", err)
	}
	active, err := systemctl.IsActive(ctx, serviceName, opts)
	if err != nil {
		return fmt.Errorf("check runner service: %w", err)
	}
	if !active {
		if err := systemctl.Start(ctx, serviceName, opts); err != nil {
			return fmt.Errorf("start runner service: %w", err)
		}
	}
	return nil
}

func ecrLogin(ctx context.Context, region, registry string) error {
	password := exec.CommandContext(ctx, "aws", "ecr", "get-login-password", "--region", region)
	login := exec.CommandContext(ctx, "docker", "login", "--username", "AWS", "--password-stdin", registry)
	pipe, err := password.StdoutPipe()
	if err != nil {
		return err
	}
	login.Stdin = pipe
	login.Stdout = io.Discard
	login.Stderr = os.Stderr
	if err := password.Start(); err != nil {
		return err
	}
	if err := login.Run(); err != nil {
		_ = password.Wait()
		return err
	}
	return password.Wait()
}
