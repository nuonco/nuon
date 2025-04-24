package plan

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (p *Planner) getInstallRegistryRepositoryConfig(ctx workflow.Context, installID, deployID string) (*configs.OCIRegistryRepository, error) {
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

	stateData, err := state.AsMap()
	if err != nil {
		return nil, errors.Wrap(err, "state data")
	}

	// NOTE(jm): this is mainly a relic of not having the outputs properly passed from the install sandbox, or a
	// good way of "cataloging" resources.
	repositoryStr, err := render.RenderV2("{{.nuon.sandbox.outputs.ecr.repository_url}}", stateData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render repository url")
	}

	registryURL, err := render.RenderV2("{{.nuon.sandbox.outputs.ecr.registry_url}}", stateData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render repository url")
	}

	return &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypeECR,
		Plugin:       "oci",
		Repository:   repositoryStr,
		LoginServer:  registryURL,
		Region:       stack.AWSStackOutputs.Region,
		ECRAuth: &credentials.Config{
			Region: stack.AWSStackOutputs.Region,
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:     stack.AWSStackOutputs.MaintenanceIAMRoleARN,
				SessionName: fmt.Sprintf("oci-sync-%s-%s", installID, deployID),
			},
		},
	}, nil
}

func (b *Planner) getOrgRegistryRepositoryConfig(ctx workflow.Context, installID, deployID string) (*configs.OCIRegistryRepository, error) {
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack by install id")
	}

	installStack, err := activities.AwaitGetInstallStackByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	stackOutputs, err := activities.AwaitGetInstallStackOutputs(ctx, installStack.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack outputs")
	}

	accessInfo, err := activities.AwaitGetOrgECRAccessInfo(ctx, install.OrgID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get access info")
	}

	appRepoName := fmt.Sprintf("%s/%s", install.OrgID, install.AppID)
	appRepoURI := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", accessInfo.RegistryID,
		accessInfo.Region, appRepoName)

	return &configs.OCIRegistryRepository{
		Repository:   appRepoURI,
		Region:       stackOutputs.AWSStackOutputs.Region,
		RegistryType: configs.OCIRegistryTypePrivateOCI,
		OCIAuth: &configs.OCIRegistryAuth{
			Username: accessInfo.Username,
			Password: accessInfo.RegistryToken,
		},
		LoginServer: strings.TrimPrefix(accessInfo.ServerAddress, "https://"),
	}, nil
}
