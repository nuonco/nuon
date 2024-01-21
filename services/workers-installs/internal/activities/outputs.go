package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/workflows/dal"
	"google.golang.org/protobuf/types/known/structpb"
)

type FetchSandboxOutputsRequest struct {
	OrgID     string
	AppID     string
	InstallID string
}

func (a *Activities) FetchSandboxOutputs(ctx context.Context, req FetchSandboxOutputsRequest) (*structpb.Struct, error) {
	dalClient, err := a.getDalClient(req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get dal client: %w", err)
	}

	outputs, err := dalClient.GetInstallSandboxOutputs(ctx, req.OrgID, req.AppID, req.InstallID)
	if err != nil {
		return nil, fmt.Errorf("unable to get outputs: %w", err)
	}

	return outputs, nil
}

func (s *Activities) getDalClient(orgID string) (dal.Client, error) {
	dalClient, err := dal.New(s.v,
		dal.WithSettings(dal.Settings{
			InstallsBucket:                s.cfg.InstallationsBucket,
			InstallsBucketIAMRoleTemplate: s.cfg.OrgInstallationsRoleTemplate,
			OrgsBucket:                    s.cfg.InstallationsBucket,
		}),
		dal.WithOrgID(orgID))
	if err != nil {
		return nil, fmt.Errorf("unable to get dal client: %w", err)
	}

	return dalClient, nil
}
