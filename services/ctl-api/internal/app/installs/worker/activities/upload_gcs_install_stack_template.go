package activities

import (
	"context"
	"fmt"

	gcs "cloud.google.com/go/storage"
)

type UploadGCSInstallStackTemplateRequest struct {
	Bucket   string `validate:"required"`
	Key      string `validate:"required"`
	Template []byte `validate:"required"`
}

type UploadGCSInstallStackTemplateResponse struct {
	URL string
}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 2m
// @start-to-close-timeout 2m
// @max-retries 3
func (a *Activities) UploadGCSInstallStackTemplate(ctx context.Context, req *UploadGCSInstallStackTemplateRequest) (*UploadGCSInstallStackTemplateResponse, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	client, err := gcs.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create GCS client: %w", err)
	}
	defer client.Close()

	writer := client.Bucket(req.Bucket).Object(req.Key).NewWriter(ctx)
	writer.ContentType = "application/json"
	if _, err := writer.Write(req.Template); err != nil {
		return nil, fmt.Errorf("unable to write to GCS: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("unable to close GCS writer: %w", err)
	}

	// The bucket is configured with public read access (allUsers: objectViewer).
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", req.Bucket, req.Key)

	return &UploadGCSInstallStackTemplateResponse{URL: url}, nil
}
