package dal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/powertoolsdev/mono/pkg/aws/s3downloader"
	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
)

// GetAppProvisionRequest returns a provision request for an app
func (r *repo) GetAppProvisionRequest(ctx context.Context, orgID, appID string) (*appsv1.ProvisionRequest, error) {
	creds := r.getAppsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.AppsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.AppPath(orgID, appID), requestFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	req, err := unmarshalRequest(byts)
	if err != nil {
		return nil, fmt.Errorf("unable to get apps provision request: %w", err)
	}

	appReq := req.Request.GetAppProvision()
	if appReq == nil {
		return nil, fmt.Errorf("app request not set on shared request")
	}

	return appReq, nil
}

func (r *repo) GetAppProvisionResponse(ctx context.Context, orgID, appID string) (*appsv1.ProvisionResponse, error) {
	creds := r.getAppsCredentials(ctx)
	client, err := s3downloader.New(r.Settings.AppsBucket, s3downloader.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	key := filepath.Join(prefix.AppPath(orgID, appID), responseFilename)
	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	resp, err := unmarshalResponse(byts)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal shared response: %w", err)
	}

	appResp := resp.Response.GetAppsProvision()
	if appResp == nil {
		return nil, fmt.Errorf("app response not set on shared response")
	}

	return appResp, nil
}
