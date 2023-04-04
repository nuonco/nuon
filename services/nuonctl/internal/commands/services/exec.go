package services

import (
	"context"
	"fmt"

	eksclient "github.com/powertoolsdev/mono/pkg/aws/eks-client"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/services/command"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/services/config"
	"k8s.io/client-go/kubernetes"
)

var DEFAULT_ENV_VARS map[string]string = map[string]string{
	"TEMPORAL_HOST": "temporal:7233",
	"HOST_IP":       "0.0.0.0",
}

func (c *commands) Exec(ctx context.Context, svcName string, cmd []string) error {
	if len(cmd) < 1 {
		return fmt.Errorf("must pass at least one command to execute")
	}
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

	return c.exec(ctx, svcName, cfg, env, cmd)
}

// run a service locally with the provided environment + cfg using earthly
func (c *commands) exec(ctx context.Context, svc string, cfg *config.Config, env map[string]string, args []string) error {
	for k, v := range DEFAULT_ENV_VARS {
		env[k] = v
	}
	for k, v := range cfg.Env {
		env[k] = v
	}

	cmd, err := command.New(c.v,
		command.WithEnv(env),
		command.WithCmd(args[0]),
		command.WithArgs(args[0:]),
	)
	if err != nil {
		return fmt.Errorf("unable to build command: %w", err)
	}

	if err := cmd.Exec(ctx); err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}
	return nil
}
