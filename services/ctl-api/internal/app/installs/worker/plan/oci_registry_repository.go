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

	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) getInstallRegistryRepositoryConfig(
	ctx workflow.Context,
	installDeploy *app.InstallDeploy,
	compBuild *app.ComponentBuild,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	installState *state.State,
	roleSelection *operationroles.RoleSelection,
) (*configs.OCIRegistryRepository, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	stateData, err := installState.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "state data")
	}

	sessionName := fmt.Sprintf("oci-sync-%s-%s", installDeploy.InstallID, installDeploy.ID)
	cloudAuth, err := p.getAuthForDeploy(ctx, roleSelection, stack, sessionName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for install registry")
	}

	cfg := &configs.OCIRegistryRepository{
		Plugin: "oci",
	}

	// NOTE(jm): this is mainly a relic of not having the outputs properly passed from the install sandbox, or a
	// good way of "cataloging" resources.
	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:

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
		cfg.Region = stack.InstallStackOutputs.AWSStackOutputs.Region
		cfg.ECRAuth = cloudAuth.AWS

	case stack.InstallStackOutputs.AzureStackOutputs != nil:

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
		// Per-component paths so resolved-version tags can't collide across
		// components. ACR creates nested repositories implicitly on push.
		cfg.Repository = repositoryStr + "/" + imageNameSegment(installDeploy.ComponentName)
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

	case stack.InstallStackOutputs.GCPStackOutputs != nil:

		cfg.RegistryType = configs.OCIRegistryTypeGAR
		repositoryStr, err := render.RenderV2("{{.nuon.sandbox.outputs.gar.repository_url}}", stateData)
		if err != nil {
			l.Error("error rendering repository",
				zap.Any("repository", repositoryStr),
				zap.Error(err),
				zap.Any("state", stateData),
			)
			return nil, errors.Wrap(err, "unable to render gar repository url")
		}
		// GAR requires an image name within the repo: HOST/PROJECT/REPO/IMAGE.
		// Per-component paths so resolved-version tags can't collide across components.
		cfg.Repository = repositoryStr + "/" + imageNameSegment(installDeploy.ComponentName)
		loginServer, err := render.RenderV2("{{.nuon.sandbox.outputs.gar.registry_url}}", stateData)
		if err != nil {
			l.Error("error rendering registy url",
				zap.Any("registry-url", loginServer),
				zap.Error(err),
				zap.Any("state", stateData),
			)
			return nil, errors.Wrap(err, "unable to render gar login server")
		}
		cfg.LoginServer = loginServer
		cfg.Region = stack.InstallStackOutputs.GCPStackOutputs.Region
	}

	return cfg, nil
}

// imageNameSegment reduces a component name to a docker image path segment /
// tag prefix: lowercase, every run of non-alphanumerics (including "_")
// collapsed to a single "-", no leading or trailing separator.
func imageNameSegment(componentName string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(componentName) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteRune('-')
			lastDash = true
		}
	}
	segment := strings.Trim(b.String(), "-")
	if segment == "" {
		return "app"
	}
	return segment
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
	loginServer := strings.TrimPrefix(accessInfo.ServerAddress, "https://")

	// For GCP/GAR, the RegistryID from GetOrgECRAccessInfo contains the full GAR URL
	// (e.g. "us-central1-docker.pkg.dev/project/repo"). Use it to build the full image path.
	// Always use PrivateOCI with static credentials — the install runner may not have GCP
	// default credentials (it runs in the customer's cloud, not ours).
	if accessInfo.RegistryID != "" && strings.Contains(accessInfo.ServerAddress, "pkg.dev") {
		garURL := accessInfo.RegistryID
		if idx := strings.Index(garURL, "/"); idx != -1 {
			loginServer = garURL[:idx]
			appRepoName = fmt.Sprintf("%s/%s/%s", garURL[idx+1:], install.OrgID, install.AppID)
		}
	}

	return &configs.OCIRegistryRepository{
		Repository:   appRepoName,
		Region:       "",
		RegistryType: configs.OCIRegistryTypePrivateOCI,
		OCIAuth: &configs.OCIRegistryAuth{
			Username: accessInfo.Username,
			Password: accessInfo.RegistryToken,
		},
		LoginServer: loginServer,
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
