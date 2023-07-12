package activities

import (
	"context"
	"fmt"
	"time"

	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	vcsv1 "github.com/powertoolsdev/mono/pkg/types/components/vcs/v1"
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

type Activities struct {
	db                     *gorm.DB
	githubAppID            string
	githubAppKeySecretName string
	componentService       services.ComponentService
	componentRepo          repos.ComponentRepo
}

func New(db *gorm.DB, githubAppID, githubAppKeySecretName string) *Activities {
	// TODO fix up logger here
	componentService := services.NewComponentService(db, zap.L())

	// The component service doesn't allow updating an existing componet,
	// so we need to create an instance of the component repo.
	// The service has one, but doesn't expose it.
	componentRepo := repos.NewComponentRepo(db)

	return &Activities{
		db:                     db,
		githubAppID:            githubAppID,
		githubAppKeySecretName: githubAppKeySecretName,
		componentService:       componentService,
		componentRepo:          componentRepo,
	}
}

func (a *Activities) CreatePlanRequest(ctx context.Context, req *workflowbuildv1.BuildRequest) (*planv1.CreatePlanRequest, error) {
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
func (a *Activities) addVcsConfig(component *models.Component, gitRef, ghInstallID string) (*componentv1.Component, error) {
	dbConfig := &componentv1.Component{}
	unmarshaller := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if err := unmarshaller.Unmarshal([]byte(component.Config.String()), dbConfig); err != nil {
		return nil, err
	}

	if dbConfig.BuildCfg == nil {
		return nil, fmt.Errorf("build config was nil")
	}

	vcsCfg := dbConfig.BuildCfg.GetVcsCfg()
	if vcsCfg != nil {
		switch vcs := vcsCfg.Cfg.(type) {
		case *vcsv1.Config_ConnectedGithubConfig:
			vcs.ConnectedGithubConfig.GitRef = gitRef
			vcs.ConnectedGithubConfig.GithubAppKeyId = a.githubAppID
			vcs.ConnectedGithubConfig.GithubAppKeySecretName = a.githubAppKeySecretName
			vcs.ConnectedGithubConfig.GithubInstallId = ghInstallID
		case *vcsv1.Config_PublicGitConfig:
			vcs.PublicGitConfig.GitRef = gitRef
		default:
		}
	}
	dbConfig.DeployCfg.Timeout = durationpb.New(defaultBuildTimeout)

	return dbConfig, nil
}
