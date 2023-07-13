package dal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/mono/pkg/aws/s3downloader"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
)

// GetOrgSignupRequest returns a provision request for an org
func (r *client) GetOrgProvisionRequest(ctx context.Context, orgID string) (*orgsv1.SignupRequest, error) {
	creds := r.orgsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.OrgsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.OrgPath(orgID), requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	req, err := unmarshalRequest(byts)
	if err != nil {
		return nil, fmt.Errorf("unable to get orgs provision request: %w", err)
	}

	orgReq := req.Request.GetOrgSignup()
	if orgReq == nil {
		return nil, fmt.Errorf("org request not set on shared request")
	}

	return orgReq, nil
}

func (r *client) GetOrgProvisionResponse(ctx context.Context, orgID string) (*orgsv1.SignupResponse, error) {
	creds := r.orgsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.OrgsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.OrgPath(orgID), responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	resp, err := unmarshalResponse(byts)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal shared response: %w", err)
	}

	orgResp := resp.Response.GetOrgSignup()
	if orgResp == nil {
		return nil, fmt.Errorf("org response not set on shared response")
	}

	return orgResp, nil
}
