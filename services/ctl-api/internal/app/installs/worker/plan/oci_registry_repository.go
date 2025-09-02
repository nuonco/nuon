package plan

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) getInstallRegistryRepositoryConfig(ctx workflow.Context, installID, deployID string) (*configs.OCIRegistryRepository, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	installStack, err := activities.AwaitGetInstallStackByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	state, err := activities.AwaitGetInstallStateByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}

	stack, err := activities.AwaitGetInstallStackOutputs(ctx, installStack.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack outputs")
	}

	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "state data")
	}

	cfg := &configs.OCIRegistryRepository{
		RegistryType: "",
		Plugin:       "oci",
		Repository:   "",
		LoginServer:  "",
		Region:       "",
		ECRAuth:      nil,
		ACRAuth:      nil,
	}

	// NOTE(jm): this is mainly a relic of not having the outputs properly passed from the install sandbox, or a
	// good way of "cataloging" resources.
	switch {
	case stack.AWSStackOutputs != nil:

		cfg.RegistryType = configs.OCIRegistryTypeECR
		repositoryStr, err := render.RenderV2("{{.nuon.sandbox.outputs.ecr.repository_url}}", stateData)
		if err != nil {
			l.Error("error rendering repository",
				zap.Any("repository", repositoryStr),
				zap.Error(err),
				zap.Any("state", stateData),
			)
			return nil, errors.Wrap(err, "unable to render ecr repository url")
		}
		cfg.Repository = repositoryStr
		loginServer, err := render.RenderV2("{{.nuon.sandbox.outputs.ecr.registry_url}}", stateData)
		if err != nil {
			l.Error("error rendering registy url",
				zap.Any("registry-url", loginServer),
				zap.Error(err),
				zap.Any("state", stateData),
			)
			return nil, errors.Wrap(err, "unable to render acr login server")
		}
		cfg.LoginServer = loginServer
		cfg.Region = stack.AWSStackOutputs.Region
		cfg.ECRAuth = &credentials.Config{
			Region: stack.AWSStackOutputs.Region,
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:     stack.AWSStackOutputs.MaintenanceIAMRoleARN,
				SessionName: fmt.Sprintf("oci-sync-%s-%s", installID, deployID),
			},
		}

	case stack.AzureStackOutputs != nil:

		cfg.RegistryType = configs.OCIRegistryTypeACR
		repositoryStr, err := render.RenderV2("{{.nuon.sandbox.outputs.acr.name}}", stateData)
		if err != nil {
			l.Error("error rendering repository",
				zap.Any("repository", repositoryStr),
				zap.Error(err),
				zap.Any("state", stateData),
			)
			return nil, errors.Wrap(err, "unable to render acr repository name")
		}
		cfg.Repository = repositoryStr
		loginServer, err := render.RenderV2("{{.nuon.sandbox.outputs.acr.login_server}}", stateData)
		if err != nil {
			l.Error("error rendering registy url",
				zap.Any("registry-url", loginServer),
				zap.Error(err),
				zap.Any("state", stateData),
			)
			return nil, errors.Wrap(err, "unable to render acr login server")
		}
		cfg.LoginServer = loginServer
		cfg.ACRAuth = &azurecredentials.Config{
			UseDefault: true,
		}
	}

	return cfg, nil
}

func (b *Planner) getOrgRegistryRepositoryConfig(ctx workflow.Context, installID, deployID string) (*configs.OCIRegistryRepository, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack by install id")
	}

	var accessInfo *activities.OrgECRAccessInfo
	if install.Org.SandboxMode {
		l.Info("sandbox-mode enabled, creating fake access info")
		accessInfo = generics.GetFakeObj[*activities.OrgECRAccessInfo]()
	} else {
		accessInfo, err = activities.AwaitGetOrgECRAccessInfo(ctx, install.OrgID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get access info")
		}
	}

	appRepoName := fmt.Sprintf("%s/%s", install.OrgID, install.AppID)
	return &configs.OCIRegistryRepository{
		Repository:   appRepoName,
		Region:       "",
		RegistryType: configs.OCIRegistryTypePrivateOCI,
		OCIAuth: &configs.OCIRegistryAuth{
			Username: accessInfo.Username,
			Password: accessInfo.RegistryToken,
		},
		LoginServer: strings.TrimPrefix(accessInfo.ServerAddress, "https://"),
	}, nil
}

// RenderText does the same thing as render.RenderV2, but using "text/template" instead of "html/template",
// to avoid escaping special characters.
func RenderText(inputVal string, data map[string]interface{}) (string, error) {
	data = render.EnsurePrefix(data)

	if !strings.Contains(inputVal, ".nuon") {
		return inputVal, nil
	}

	funcMap := template.FuncMap{
		"now": time.Now,
	}

	temp, err := template.New("input").
		Funcs(funcMap).
		Funcs(sprig.FuncMap()).
		Option("missingkey=error").
		Parse(inputVal)
	if err != nil {
		return inputVal, err
	}

	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, data); err != nil {
		return inputVal, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.String(), nil
}
