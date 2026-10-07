package ecrrepository

import (
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type DeprovisionECRRepositoryRequest struct {
	OrgID string
	AppID string
}

type DeprovisionECRRepositoryResponse struct{}

// TODO: rename this package and its workflows to be cloud agnostic; they also handle GAR and ACR.
//
// @temporal-gen-v2 workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{.CallerID}}-deprovision-ecr-repo
func (w Wkflow) DeprovisionECRRepository(ctx workflow.Context, req *DeprovisionECRRepositoryRequest) (*DeprovisionECRRepositoryResponse, error) {
	l := log.With(workflow.GetLogger(ctx))

	var (
		cloud string
		err   error
	)

	switch {
	case w.Cfg.IsGCP():
		cloud = "gcp"
		l.Debug("destroying gar package")
		_, err = AwaitDeleteGARPackage(ctx, &DeleteGARPackageRequest{
			OrgID: req.OrgID,
			AppID: req.AppID,
		})
	case w.Cfg.IsAzure():
		cloud = "azure"
		l.Debug("destroying acr repository")
		_, err = AwaitDeleteACRRepository(ctx, &DeleteACRRepositoryRequest{
			OrgID: req.OrgID,
			AppID: req.AppID,
		})
	default:
		cloud = "aws"
		l.Debug("destroying ecr repository")
		_, err = AwaitDeleteRepository(ctx, &DeleteRepositoryRequest{
			OrgID: req.OrgID,
			AppID: req.AppID,
		})
	}
	if err != nil {
		l.Error("unable to delete app image repository",
			zap.Error(err),
			zap.String("cloud", cloud),
			zap.String("org_id", req.OrgID),
			zap.String("app_id", req.AppID),
		)
	}

	return &DeprovisionECRRepositoryResponse{}, nil
}
