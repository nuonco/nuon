package workflows

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	assumerole "github.com/powertoolsdev/mono/pkg/aws-assume-role"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
)

// TODO(jm): this should use s3downloader properly
func (r *repo) GetInstallProvisionRequest(ctx context.Context, installID string) (*installsv1.ProvisionRequest, error) {
	assumer, err := assumerole.New(r.v,
		assumerole.WithRoleARN(r.IAMRoleARN),
		assumerole.WithRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to assume role: %w", err)
	}

	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return nil, err
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
			return nil, err2
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
		return nil, fmt.Errorf("unable to find previous request for install")
	}

	// grab the object
	objReq := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &key,
	}
	objResp, err := client.GetObject(ctx, objReq)
	if err != nil {
		return nil, err
	}
	byts, err := io.ReadAll(objResp.Body)
	if err != nil {
		return nil, err
	}

	resp, err := unmarshalRequest(byts)
	if err != nil {
		return nil, err
	}

	if resp.Request.GetInstallProvision() == nil {
		return nil, fmt.Errorf("invalid request object")
	}

	return resp.Request.GetInstallProvision(), nil
}
