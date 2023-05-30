package activities

import (
	"context"
	"fmt"
	"time"

	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	workflowbuildv1 "github.com/powertoolsdev/mono/pkg/types/workflows/builds/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/services"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

const (
	defaultBuildTimeout time.Duration = time.Minute * 5
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

func (a *activities) CreatePlanRequest(ctx context.Context, req *workflowbuildv1.BuildRequest) (*planv1.CreatePlanRequest, error) {
	// TODO fix up logger here
	component, err := a.componentService.GetComponent(ctx, req.ComponentId)
	if err != nil {
		return nil, fmt.Errorf("error during StartBuild.CreatePlan loading component: %w", err)
	}

	dbConfig, err := a.addVcsConfig(component, req.GitRef, component.App.Org.GithubInstallID)
	if err != nil {
		return nil, err
	}

	// create plan for container build
	planReq := &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				Type:         planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_BUILD,
				OrgId:        component.App.OrgID,
				AppId:        component.AppID,
				DeploymentId: req.BuildId,
				Component: &componentv1.Component{
					Id:        component.ID,
					BuildCfg:  dbConfig.BuildCfg,
					DeployCfg: dbConfig.DeployCfg,
				},
			},
		},
	}

	return planReq, nil
}

// AddVcsConfig populates the component's build config with vcs config data, if the component has a build config.
func (a *activities) addVcsConfig(component *models.Component, gitRef, ghInstallID string) (*componentv1.Component, error) {
	dbConfig := &componentv1.Component{}
	if err := protojson.Unmarshal([]byte(component.Config.String()), dbConfig); err != nil {
		return nil, err
	}

	if dbConfig.BuildCfg != nil {
		isPrivate := false
		dockerCfg := dbConfig.BuildCfg.GetDockerCfg()
		if dockerCfg != nil {
			vcsCfg := dockerCfg.GetVcsCfg()
			connectedGithubConfig := vcsCfg.GetConnectedGithubConfig()
			if connectedGithubConfig != nil {
				isPrivate = true
			}
		}

		if isPrivate {
			// for private repos set: git_ref, github_app_key_id, github_app_key_secret_name, github_install_id
			dbConfig.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GitRef = gitRef
			dbConfig.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GithubAppKeyId = a.githubAppID
			dbConfig.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GithubAppKeySecretName = a.githubAppKeySecretName
			dbConfig.BuildCfg.GetDockerCfg().VcsCfg.GetConnectedGithubConfig().GithubInstallId = ghInstallID
		} else {
			// for public repos set only the git_ref
			dbConfig.BuildCfg.GetDockerCfg().VcsCfg.GetPublicGitConfig().GitRef = gitRef
		}
	}

	dbConfig.DeployCfg.Timeout = durationpb.New(defaultBuildTimeout)

	return dbConfig, nil
}
