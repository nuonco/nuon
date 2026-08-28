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

	return p.RenderInstallRegistryRepository(l, &RenderInstallRegistryRepositoryInput{
		InstallDeploy: installDeploy,
		Stack:         stack,
		StateData:     stateData,
		CloudAuth:     cloudAuth,
	})
}

// RenderInstallRegistryRepositoryInput carries the already-loaded data an install registry repository is rendered from.
type RenderInstallRegistryRepositoryInput struct {
	InstallDeploy *app.InstallDeploy
	Stack         *app.InstallStack
	StateData     map[string]any
	CloudAuth     *CloudAuth
}

// RenderInstallRegistryRepository renders an install registry repository from already-loaded inputs.
func (p *Planner) RenderInstallRegistryRepository(
	l *zap.Logger,
	in *RenderInstallRegistryRepositoryInput,
) (*configs.OCIRegistryRepository, error) {
	cfg := &configs.OCIRegistryRepository{
		Plugin: "oci",
	}

	// NOTE(jm): this is mainly a relic of not having the outputs properly passed from the install sandbox, or a
	// good way of "cataloging" resources.
	switch {
	case in.Stack.InstallStackOutputs.AWSStackOutputs != nil:

		cfg.RegistryType = configs.OCIRegistryTypeECR
		repositoryStr, err := render.RenderV2("{{.nuon.sandbox.outputs.ecr.repository_url}}", in.StateData)
		if err != nil {
			l.Error("error rendering repository",
				zap.Any("repository", repositoryStr),
				zap.Error(err),
				zap.Any("state", in.StateData),
			)
			return nil, errors.Wrap(err, "unable to render ecr repository url")
		}
		cfg.Repository = repositoryStr
		loginServer, err := render.RenderV2("{{.nuon.sandbox.outputs.ecr.registry_url}}", in.StateData)
		if err != nil {
			l.Error("error rendering registy url",
				zap.Any("registry-url", loginServer),
				zap.Error(err),
				zap.Any("state", in.StateData),
			)
			return nil, errors.Wrap(err, "unable to render acr login server")
		}
		cfg.LoginServer = loginServer
		cfg.Region = in.Stack.InstallStackOutputs.AWSStackOutputs.Region
		cfg.ECRAuth = in.CloudAuth.AWS

	case in.Stack.InstallStackOutputs.AzureStackOutputs != nil:

		cfg.RegistryType = configs.OCIRegistryTypeACR
		repositoryStr, err := render.RenderV2("{{.nuon.sandbox.outputs.acr.name}}", in.StateData)
		if err != nil {
			l.Error("error rendering repository",
				zap.Any("repository", repositoryStr),
				zap.Error(err),
				zap.Any("state", in.StateData),
			)
			return nil, errors.Wrap(err, "unable to render acr repository name")
		}
		// Per-component paths so resolved-version tags can't collide across
		// components. ACR creates nested repositories implicitly on push.
		cfg.Repository = repositoryStr + "/" + imageNameSegment(in.InstallDeploy.ComponentName)
		loginServer, err := render.RenderV2("{{.nuon.sandbox.outputs.acr.login_server}}", in.StateData)
		if err != nil {
			l.Error("error rendering registy url",
				zap.Any("registry-url", loginServer),
				zap.Error(err),
				zap.Any("state", in.StateData),
			)
			return nil, errors.Wrap(err, "unable to render acr login server")
		}
		cfg.LoginServer = loginServer
		cfg.ACRAuth = &azurecredentials.Config{
			UseDefault: true,
		}

	case in.Stack.InstallStackOutputs.GCPStackOutputs != nil:

		cfg.RegistryType = configs.OCIRegistryTypeGAR
		repositoryStr, err := render.RenderV2("{{.nuon.sandbox.outputs.gar.repository_url}}", in.StateData)
		if err != nil {
			l.Error("error rendering repository",
				zap.Any("repository", repositoryStr),
				zap.Error(err),
				zap.Any("state", in.StateData),
			)
			return nil, errors.Wrap(err, "unable to render gar repository url")
		}
		// GAR requires an image name within the repo: HOST/PROJECT/REPO/IMAGE.
		// Per-component paths so resolved-version tags can't collide across components.
		cfg.Repository = repositoryStr + "/" + imageNameSegment(in.InstallDeploy.ComponentName)
		loginServer, err := render.RenderV2("{{.nuon.sandbox.outputs.gar.registry_url}}", in.StateData)
		if err != nil {
			l.Error("error rendering registy url",
				zap.Any("registry-url", loginServer),
				zap.Error(err),
				zap.Any("state", in.StateData),
			)
			return nil, errors.Wrap(err, "unable to render gar login server")
		}
		cfg.LoginServer = loginServer
		cfg.Region = in.Stack.InstallStackOutputs.GCPStackOutputs.Region
		if in.CloudAuth.GCP != nil {
			cfg.ServiceAccountEmail = in.CloudAuth.GCP.ImpersonateServiceAccount
		}
	}

	return cfg, nil
}

// installRegistryLoginServer returns the login server of the install's own
// registry, or "" when the sandbox emits no registry outputs. Best-effort by
// design: it is used to decide whether an image ref already points at the
// install registry, and an install whose sandbox has no registry is a perfectly
// valid host for an action that pulls a public image.
func installRegistryLoginServer(stateData map[string]interface{}, stack *app.InstallStack) string {
	var tmpl string
	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		tmpl = "{{.nuon.sandbox.outputs.ecr.registry_url}}"
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		tmpl = "{{.nuon.sandbox.outputs.acr.login_server}}"
	case stack.InstallStackOutputs.GCPStackOutputs != nil:
		tmpl = "{{.nuon.sandbox.outputs.gar.registry_url}}"
	default:
		return ""
	}

	loginServer, err := render.RenderV2(tmpl, stateData)
	if err != nil {
		return ""
	}

	return strings.TrimPrefix(strings.TrimSpace(loginServer), "https://")
}

// getInstallRegistryPullConfig builds the registry config for pulling an image
// that already lives in the install's registry, so the runner authenticates
// with the install's cloud credentials rather than attempting an anonymous
// pull. Unlike getInstallRegistryRepositoryConfig this is not tied to a
// component deploy: the repository comes from the ref the caller resolved.
func getInstallRegistryPullConfig(
	repository string,
	loginServer string,
	stack *app.InstallStack,
	cloudAuth *CloudAuth,
) *configs.OCIRegistryRepository {
	cfg := &configs.OCIRegistryRepository{
		Plugin:      "oci",
		Repository:  repository,
		LoginServer: loginServer,
	}

	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		cfg.RegistryType = configs.OCIRegistryTypeECR
		cfg.Region = stack.InstallStackOutputs.AWSStackOutputs.Region
		cfg.ECRAuth = cloudAuth.AWS
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		cfg.RegistryType = configs.OCIRegistryTypeACR
		cfg.ACRAuth = &azurecredentials.Config{
			UseDefault: true,
		}
	case stack.InstallStackOutputs.GCPStackOutputs != nil:
		cfg.RegistryType = configs.OCIRegistryTypeGAR
		cfg.Region = stack.InstallStackOutputs.GCPStackOutputs.Region
		if cloudAuth.GCP != nil {
			cfg.ServiceAccountEmail = cloudAuth.GCP.ImpersonateServiceAccount
		}
	default:
		return nil
	}

	return cfg
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

	return b.RenderOrgRegistryRepository(&RenderOrgRegistryRepositoryInput{
		OrgID:         install.OrgID,
		AppID:         install.AppID,
		ServerAddress: accessInfo.ServerAddress,
		RegistryID:    accessInfo.RegistryID,
		Username:      accessInfo.Username,
		RegistryToken: accessInfo.RegistryToken,
	}), nil
}

// RenderOrgRegistryRepositoryInput carries the already-loaded data an org registry repository is rendered from.
type RenderOrgRegistryRepositoryInput struct {
	OrgID         string
	AppID         string
	ServerAddress string
	RegistryID    string
	Username      string
	RegistryToken string
}

// RenderOrgRegistryRepository renders an org registry repository from already-loaded inputs.
func (b *Planner) RenderOrgRegistryRepository(in *RenderOrgRegistryRepositoryInput) *configs.OCIRegistryRepository {
	appRepoName := fmt.Sprintf("%s/%s", in.OrgID, in.AppID)
	loginServer := strings.TrimPrefix(in.ServerAddress, "https://")

	// For GCP/GAR, the RegistryID from GetOrgECRAccessInfo contains the full GAR URL
	// (e.g. "us-central1-docker.pkg.dev/project/repo"). Use it to build the full image path.
	// Always use PrivateOCI with static credentials — the install runner may not have GCP
	// default credentials (it runs in the customer's cloud, not ours).
	if in.RegistryID != "" && strings.Contains(in.ServerAddress, "pkg.dev") {
		garURL := in.RegistryID
		if idx := strings.Index(garURL, "/"); idx != -1 {
			loginServer = garURL[:idx]
			appRepoName = fmt.Sprintf("%s/%s/%s", garURL[idx+1:], in.OrgID, in.AppID)
		}
	}

	return &configs.OCIRegistryRepository{
		Repository:   appRepoName,
		Region:       "",
		RegistryType: configs.OCIRegistryTypePrivateOCI,
		OCIAuth: &configs.OCIRegistryAuth{
			Username: in.Username,
			Password: in.RegistryToken,
		},
		LoginServer: loginServer,
	}
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
