package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
)

func (r *repo) GetInstallProvisionRequest(ctx context.Context, installID string) (*installsv1.ProvisionRequest, error) {
	resp := &installsv1.ProvisionRequest{}

	cfg, err := aws_config.LoadDefaultConfig(ctx)
	if err != nil {
		return resp, err
	}
	client := s3.NewFromConfig(cfg)

	bucketName := "nuon-org-installations-stage"
	req := &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	}

	subStr := fmt.Sprintf("install=%s", installID)
	var key string
	for {
		s3Resp, err2 := client.ListObjectsV2(ctx, req)
		if err2 != nil {
			return resp, err2
		}

		for _, obj := range s3Resp.Contents {
			if strings.Contains(*obj.Key, subStr) && strings.HasSuffix(*obj.Key, "request.json") {
				key = *obj.Key
				break
			}
		}

		if key != "" || s3Resp.ContinuationToken == nil {
			break
		}

		req.ContinuationToken = s3Resp.ContinuationToken
	}
	if key == "" {
		return resp, fmt.Errorf("unable to find previous request for install")
	}

	// grab the object
	objReq := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &key,
	}
	objResp, err := client.GetObject(ctx, objReq)
	if err != nil {
		return resp, err
	}
	byts, err := io.ReadAll(objResp.Body)
	if err != nil {
		return resp, err
	}

	if err := json.Unmarshal(byts, &resp); err != nil {
		return resp, fmt.Errorf("unable to decode to request file: %w", err)
	}

	return resp, nil
}
