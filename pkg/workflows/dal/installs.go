package dal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/mono/pkg/aws/s3downloader"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
)

// GetInstallProvisionRequest returns a provision request for an app
func (r *client) GetInstallProvisionRequest(ctx context.Context, orgID, appID, installID string) (*installsv1.ProvisionRequest, error) {
	creds := r.installsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.InstallsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.InstallPath(orgID, appID, installID), requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	req, err := unmarshalRequest(byts)
	if err != nil {
		return nil, fmt.Errorf("unable to get installs provision request: %w", err)
	}

	appReq := req.Request.GetInstallProvision()
	if appReq == nil {
		return nil, fmt.Errorf("app request not set on shared request")
	}

	return appReq, nil
}

func (r *client) GetInstallProvisionResponse(ctx context.Context, orgID, appID, installID string) (*installsv1.ProvisionResponse, error) {
	creds := r.installsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.InstallsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.InstallPath(orgID, appID, installID), responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	resp, err := unmarshalResponse(byts)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal shared response: %w", err)
	}

	appResp := resp.Response.GetInstallProvision()
	if appResp == nil {
		return nil, fmt.Errorf("app response not set on shared response")
	}

	return appResp, nil
}
