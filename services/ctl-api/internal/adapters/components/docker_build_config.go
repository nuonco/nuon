package components

import (
	"fmt"
	"strings"

	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	variablesv1 "github.com/powertoolsdev/mono/pkg/types/components/variables/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *Adapter) toEnvVars(inputVals map[string]*string) *variablesv1.EnvVars {
	vals := make([]*variablesv1.EnvVar, 0)
	for k, v := range inputVals {
		if v == nil {
			continue
		}

		vals = append(vals, &variablesv1.EnvVar{
			Name:      k,
			Value:     *v,
			Sensitive: true,
		})
	}

	return &variablesv1.EnvVars{
		Env: vals,
	}
}

func (c *Adapter) toBuildArgs(inputArgs []string) ([]*buildv1.DockerBuildArg, error) {
	args := make([]*buildv1.DockerBuildArg, 0)
	for _, arg := range inputArgs {
		pieces := strings.SplitN(arg, "=", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("invalid docker build arg: %s", arg)
		}

		args = append(args, &buildv1.DockerBuildArg{
			Key:   pieces[0],
			Value: pieces[1],
		})
	}

	return args, nil
}

func (c *Adapter) ToDockerBuildConfig(cfg *app.DockerBuildComponentConfig, connections []app.InstallDeploy, gitRef string) (*componentv1.Component, error) {
	vcsCfg, err := c.ToVCSConfig(gitRef, cfg.PublicGitVCSConfig, cfg.ConnectedGithubVCSConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get vcs config: %w", err)
	}

	buildArgs, err := c.toBuildArgs(cfg.BuildArgs)
	if err != nil {
		return nil, fmt.Errorf("invalid build args: %w", err)
	}

	return &componentv1.Component{
		Id: cfg.ComponentConfigConnection.ComponentID,
		BuildCfg: &buildv1.Config{
			Timeout: durationpb.New(defaultBuildTimeout),
			Cfg: &buildv1.Config_DockerCfg{
				DockerCfg: &buildv1.DockerConfig{
					VcsCfg:     vcsCfg,
					Dockerfile: cfg.Dockerfile,
					Target:     cfg.Target,
					BuildArgs:  buildArgs,
					EnvVars: &variablesv1.EnvVars{
						Env: []*variablesv1.EnvVar{},
					},
				},
			},
		},
		DeployCfg:   c.toBasicDeployConfig(cfg.SyncOnly, cfg.BasicDeployConfig),
		Connections: c.toConnections(connections),
	}, nil
}
