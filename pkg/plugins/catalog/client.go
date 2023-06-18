package catalog

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

const (
	defaultECRPublicRegion string = "us-east-1"
)

// ecrpublicClient is an interface for interacting with the aws client, which we use for testing
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=client_mock.go -source=client.go -package=catalog
type ecrpublicClient interface {
	DescribeImageTags(context.Context, *ecrpublic.DescribeImageTagsInput, ...func(*ecrpublic.Options)) (*ecrpublic.DescribeImageTagsOutput, error)
}

func (c *catalog) getClient(ctx context.Context) (*ecrpublic.Client, error) {
	cfg, err := credentials.Fetch(ctx, c.Credentials)
	if err != nil {
		return nil, fmt.Errorf("unable to get aws config: %w", err)
	}

	cfg.Region = defaultECRPublicRegion
	return ecrpublic.NewFromConfig(cfg), nil
}
