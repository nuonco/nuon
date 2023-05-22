package activities

import (
	"context"
	"fmt"
	"strings"

	apibuildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	componentConfig "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	vcsv1 "github.com/powertoolsdev/mono/pkg/types/components/vcs/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
	"google.golang.org/protobuf/encoding/protojson"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/services"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

type activities struct {
	db                     *gorm.DB
	githubAppID            string
	githubAppKeySecretName string
	componentService       services.ComponentService
	componentRepo          repos.ComponentRepo
}

func New(db *gorm.DB, githubAppID, githubAppKeySecretName string) *activities {
	// TODO fix up logger here
	componentService := services.NewComponentService(db, zap.L())

	// The component service doesn't allow updating an existing componet,
	// so we need to create an instance of the component repo.
	// The service has one, but doesn't expose it.
	componentRepo := repos.NewComponentRepo(db)

	return &activities{
		db:                     db,
		githubAppID:            githubAppID,
		githubAppKeySecretName: githubAppKeySecretName,
		componentService:       componentService,
		componentRepo:          componentRepo,
	}
}

func (a *activities) CreatePlanRequest(ctx context.Context, req *apibuildv1.StartBuildRequest) (*planv1.CreatePlanRequest, error) {
	// TODO fix up logger here
	component, err := a.componentService.GetComponent(ctx, req.ComponentId)
	if err != nil {
		return nil, fmt.Errorf("error during StartBuild.CreatePlan loading component: %w", err)
	}

	// retrieve commits if the component configuration contains github info
	dbConfig := &componentConfig.Component{}
	if err = protojson.Unmarshal([]byte(component.Config.String()), dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DB JSON: %w", err)
	}
	if dbConfig.BuildCfg != nil {
		dockerCfg := dbConfig.BuildCfg.GetDockerCfg()
		if dockerCfg != nil {
			err = a.getGithubInfo(ctx, dockerCfg.GetVcsCfg(), component, req.GitRef, a.githubAppID, a.githubAppKeySecretName)
			if err != nil {
				return nil, fmt.Errorf("failed to process github repo: %w", err)
			}
		}
	}

	componentProto, err := converters.ComponentModelToProto(component)
	if err != nil {
		return nil, fmt.Errorf("error during StartBuild.CreatePlan component ToProto: %w", err)
	}
	// create plan for container build
	planReq := &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				Type:  planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_BUILD,
				OrgId: component.App.OrgID,
				AppId: component.AppID,
				// TODO adjust these structs accordingly. No deploymentid relevant to this anymore.
				DeploymentId: "dplfakeforcevalidation1234", // I think this can be empty
				Component: &componentv1.Component{
					Id:        component.ID,
					BuildCfg:  componentProto.ComponentConfig.BuildCfg,
					DeployCfg: componentProto.ComponentConfig.DeployCfg,
				},
			},
		},
	}

	return planReq, nil
}

func (a *activities) getGithubInfo(ctx context.Context, vcsCfg *vcsv1.Config, component *models.Component, gitRef, githubAppID, githubAppKeySecretName string) error {
	var repo string
	connectedGithubConfig := vcsCfg.GetConnectedGithubConfig()
	githubInstallID := "" // start with empty string which also means public git repo

	// if the component config is for a connected github repo,
	// then we have a githubInstallID and should make github API calls using it
	isPrivate := connectedGithubConfig != nil
	if isPrivate {
		// Use the githubInstallID for github API calls
		repo = connectedGithubConfig.Repo
		githubInstallID = component.App.Org.GithubInstallID
	} else {
		publicGitConfig := vcsCfg.GetPublicGitConfig()
		// TODO: temporary workaround, will refactor after the component config retro
		repo = strings.ReplaceAll(publicGitConfig.Repo, "https://github.com/", "")
		repo = strings.ReplaceAll(repo, ".git", "")
	}

	// update component config with github info
	err := a.updateComponentConfig(ctx, githubInstallID, isPrivate, component, gitRef, githubAppID, githubAppKeySecretName)
	if err != nil {
		return err
	}

	return nil
}

func (a *activities) updateComponentConfig(ctx context.Context, ghInstallID string, isPrivate bool, component *models.Component, gitRef, githubAppID, githubAppKeySecretName string) error {
	// unmarshal existing JSON
	dbConfig, _ := component.Config.MarshalJSON()
	componentConfiguration := &componentConfig.Component{}
	if err := protojson.Unmarshal(dbConfig, componentConfiguration); err != nil {
		return err
	}

	if isPrivate {
		// for private repos set: git_ref, github_app_key_id, github_app_key_secret_name, github_install_id
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GitRef = gitRef
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GithubAppKeyId = githubAppID
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GithubAppKeySecretName = githubAppKeySecretName
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GithubInstallId = ghInstallID
	} else {
		// for public repos set only the git_ref
		componentConfiguration.BuildCfg.GetDockerCfg().VcsCfg.GetPublicGitConfig().GitRef = gitRef
	}

	//marshal JSON
	var updatedConfig []byte
	updatedConfig, err := protojson.Marshal(componentConfiguration)
	if err != nil {
		return err
	}
	component.Config = updatedConfig

	// update component record with new configuration
	component, err = a.componentRepo.Update(ctx, component)
	if err != nil {
		return err
	}

	return nil
}
