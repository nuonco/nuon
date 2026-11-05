package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	ecrauthorization "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/pkg/aws/ecr/sessionpolicy"
	"github.com/nuonco/nuon/pkg/azure/acr"
)

type GetOrgECRAccessInfoRequest struct {
	OrgID string `validate:"required"`
	AppID string `validate:"required"`
}

type OrgECRAccessInfo struct {
	RegistryID    string
	Username      string
	RegistryToken string
	ServerAddress string
	Region        string
}

// @temporal-gen-v2 activity
func (a *Activities) GetOrgECRAccessInfo(ctx context.Context, req *GetOrgECRAccessInfoRequest) (*OrgECRAccessInfo, error) {
	if req == nil || req.OrgID == "" || req.AppID == "" {
		return nil, fmt.Errorf("org id and app id are required")
	}
	if a.cfg.IsGCP() {
		return a.getOrgGARAccessInfo(ctx)
	}
	if a.cfg.IsAzure() {
		return a.getOrgACRAccessInfo(ctx, req)
	}

	repositoryARN, err := sessionpolicy.RepositoryARNFromParts(a.cfg.AppRegion, a.cfg.ManagementECRRegistryID, req.OrgID+"/"+req.AppID)
	if err != nil {
		return nil, fmt.Errorf("unable to build ecr repository arn: %w", err)
	}
	policy, err := sessionpolicy.Document(repositoryARN, true)
	if err != nil {
		return nil, err
	}

	ecr, err := ecrauthorization.New(a.v,
		ecrauthorization.WithCredentials(&credentials.Config{
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:       a.cfg.ManagementIAMRoleARN,
				SessionName:   fmt.Sprintf("oci-sync-%s", req.OrgID),
				SessionPolicy: policy,
			},
		}),
		ecrauthorization.WithRegistryID(a.cfg.ManagementECRRegistryID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create ecrauthorizer for image sync: %w", err)
	}

	ecrAuth, err := ecr.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization: %w", err)
	}

	return &OrgECRAccessInfo{
		RegistryID:    a.cfg.ManagementECRRegistryID,
		Username:      ecrAuth.Username,
		RegistryToken: ecrAuth.RegistryToken,
		ServerAddress: ecrAuth.ServerAddress,
	}, nil
}

// getOrgGARAccessInfo refuses to mint a token. Artifact Registry docker login
// only accepts a cloud-platform token, and there is no reader identity to
// impersonate. The pod token is the control plane identity.
func (a *Activities) getOrgGARAccessInfo(context.Context) (*OrgECRAccessInfo, error) {
	return nil, fmt.Errorf("refusing to mint a control plane token for artifact registry")
}

// getOrgACRAccessInfo returns an ACR access token scoped to this app's repository.
func (a *Activities) getOrgACRAccessInfo(ctx context.Context, req *GetOrgECRAccessInfoRequest) (*OrgECRAccessInfo, error) {
	acrService := a.cfg.ManagementACRRegistryURL

	refreshToken, err := acr.GetRepositoryToken(ctx, nil, acrService, zap.L())
	if err != nil {
		return nil, fmt.Errorf("unable to get ACR repository token: %w", err)
	}
	token, err := acr.AccessToken(ctx, acrService, "repository:"+req.OrgID+"/"+req.AppID+":pull,push", refreshToken)
	if err != nil {
		return nil, fmt.Errorf("unable to scope ACR access token: %w", err)
	}

	return &OrgECRAccessInfo{
		RegistryID:    acrService,
		Username:      acr.DefaultACRUsername,
		RegistryToken: token,
		ServerAddress: "https://" + acrService,
	}, nil
}
