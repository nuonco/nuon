package eks_client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=cluster_mock_test.go -source=cluster.go -package=eks_client
func (e *eksClient) GetCluster(ctx context.Context) (*ekstypes.Cluster, error) {
	cfg, err := e.getConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get config: %w", err)
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
		return nil, fmt.Errorf("unable to get describe cluster: %w", err)
	}

	return resp.Cluster, nil
}

type awsEKSClient interface {
	DescribeCluster(context.Context, *eks.DescribeClusterInput, ...func(*eks.Options)) (*eks.DescribeClusterOutput, error)
}
