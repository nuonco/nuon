package eksclient

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=cluster_mock_test.go -source=cluster.go -package=eksclient
func (e *eksClient) GetCluster(ctx context.Context) (*ekstypes.Cluster, error) {
	cfg, err := credentials.Fetch(ctx, e.AWSAuth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get aws credentials")
	}

	eksClient := eks.NewFromConfig(cfg)
	cluster, err := e.getCluster(ctx, eksClient)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster %s: %w", e.ClusterName, err)
	}

	return cluster, nil
}

func (e *eksClient) getCluster(ctx context.Context, client awsEKSClient) (*ekstypes.Cluster, error) {
	req := &eks.DescribeClusterInput{
		Name: generics.ToPtr(e.ClusterName),
	}

	resp, err := client.DescribeCluster(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to describe cluster: %w", err)
	}

	return resp.Cluster, nil
}

type awsEKSClient interface {
	DescribeCluster(context.Context, *eks.DescribeClusterInput, ...func(*eks.Options)) (*eks.DescribeClusterOutput, error)
}
