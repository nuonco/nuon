package services

import (
	"context"
	"fmt"

	eksclient "github.com/powertoolsdev/mono/pkg/aws/eks-client"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/services/command"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/services/config"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultLocalImageRepoTmpl = "nuon.co/%s"
	defaultLocalImageTag      = "latest"
)

func (c *commands) Run(ctx context.Context, svcName string, args []string) error {
	root, err := rootDir()
	if err != nil {
		return fmt.Errorf("unable to get root dir: %w", err)
	}

	cfgLoader, err := config.New(c.v, config.WithRootDir(root),
		config.WithService(svcName))
	if err != nil {
		return fmt.Errorf("unable to create config loader: %w", err)
	}

	cfg, err := cfgLoader.Load()
	if err != nil {
		return fmt.Errorf("unable to load config loader: %w", err)
	}

	if cfg.Type == config.ServiceTypeBinary {
		return fmt.Errorf("unable to run binary services")
	}

	// load environment
	eksClienter, err := eksclient.New(c.v,
		eksclient.WithClusterName(clusterName),
		eksclient.WithRegion(region),
		eksclient.WithRoleSessionName(assumeRoleSessionName),
		eksclient.WithRoleARN(assumeRoleARN),
	)
	if err != nil {
		return fmt.Errorf("unable to get eks client creator: %w", err)
	}

	kubeCfg, err := eksClienter.GetKubeConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to get kube config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return fmt.Errorf("failed to create kube client: %w", err)
	}

	configMapsClient := kubeClient.CoreV1().ConfigMaps(serviceNamespace)
	env, err := c.getServiceEnv(ctx, configMapsClient, svcName)
	if err != nil {
		return fmt.Errorf("unable to get config map: %w", err)
	}

	imageURL, err := c.buildService(ctx, svcName)
	if err != nil {
		return fmt.Errorf("unable to build service %s: %w", svcName, err)
	}

	return c.runService(ctx, svcName, imageURL, cfg, env, args)
}

// run a service locally with the provided environment + cfg using earthly
func (c *commands) runService(ctx context.Context, svcName, imageURL string, cfg *config.Config, env map[string]string, userArgs []string) error {
	args := []string{
		"run",
		"--interactive",
		"--tty",
		"--rm",
		"-l",
		"nuonctl",
		"--name=" + svcName,
		"--network=mono_default",
	}
	for k, v := range DefaultEnvVars {
		env[k] = v
	}
	for k, v := range cfg.Env {
		env[k] = v
	}
	for k, v := range env {
		args = append(args, "--env", fmt.Sprintf("%s=%v", k, v))
	}

	if cfg.Port > 0 {
		args = append(args, "--publish", fmt.Sprintf("%d", cfg.Port))
	}

	args = append(args, imageURL)
	args = append(args, userArgs...)

	cmd, err := command.New(c.v,
		command.WithEnv(map[string]string{}),
		command.WithCmd("docker"),
		command.WithArgs(args),
	)
	if err != nil {
		return fmt.Errorf("unable to create run command: %w", err)
	}
	if err := cmd.Exec(ctx); err != nil {
		return fmt.Errorf("unable to run service: %w", err)
	}

	return nil
}

// run a service locally with the provided environment + cfg using earthly
func (c *commands) buildService(ctx context.Context, svc string) (string, error) {
	svcDir, err := serviceDir(svc)
	if err != nil {
		return "", fmt.Errorf("unable to get service dir: %w", err)
	}

	imageRepo := fmt.Sprintf(defaultLocalImageRepoTmpl, svc)
	cmd, err := command.New(c.v,
		command.WithEnv(map[string]string{}),
		command.WithCmd("earthly"),
		command.WithCwd(svcDir),
		command.WithArgs([]string{
			"--output",
			"--ci",
			"+docker",
			"--image_tag=" + defaultLocalImageTag,
			"--repo=" + imageRepo,
		}),
	)
	if err != nil {
		return "", fmt.Errorf("unable to create build command: %w", err)
	}
	if err := cmd.Exec(ctx); err != nil {
		return "", fmt.Errorf("unable to build service: %w", err)
	}

	return fmt.Sprintf("%s:%s", imageRepo, defaultLocalImageTag), nil
}
