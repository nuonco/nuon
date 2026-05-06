package stack

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
)

const eksVersion = "1.30"

// CreateEKSCluster creates the control plane and waits for ACTIVE.
func CreateEKSCluster(ctx context.Context, log *slog.Logger, c *eks.Client, iamc *iam.Client, st *State) error {
	if st.ClusterName == "" {
		st.ClusterName = "nuon-" + st.InstallID
	}
	roleArn, err := getRoleARN(ctx, iamc, st.ClusterRoleName)
	if err != nil {
		return err
	}

	log = log.With("cluster", st.ClusterName)

	subnets := append([]string{}, st.PrivateSubnetIDs...)
	subnets = append(subnets, st.PublicSubnetIDs...)

	_, err = c.CreateCluster(ctx, &eks.CreateClusterInput{
		Name:    &st.ClusterName,
		Version: aws.String(eksVersion),
		RoleArn: &roleArn,
		ResourcesVpcConfig: &ekstypes.VpcConfigRequest{
			SubnetIds: subnets,
		},
	})
	if err != nil {
		var ae smithy.APIError
		if !(errors.As(err, &ae) && ae.ErrorCode() == "ResourceInUseException") {
			return fmt.Errorf("create cluster: %w", err)
		}
		log.Info("cluster exists, reusing")
	} else {
		log.Info("create cluster initiated, waiting for ACTIVE")
	}
	if err := st.Save(); err != nil {
		return err
	}

	deadline := time.Now().Add(30 * time.Minute)
	for time.Now().Before(deadline) {
		out, err := c.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: &st.ClusterName})
		if err != nil {
			return fmt.Errorf("describe cluster: %w", err)
		}
		switch out.Cluster.Status {
		case ekstypes.ClusterStatusActive:
			st.ClusterEndpoint = aws.ToString(out.Cluster.Endpoint)
			if out.Cluster.CertificateAuthority != nil {
				st.ClusterCAData = aws.ToString(out.Cluster.CertificateAuthority.Data)
			}
			if out.Cluster.Identity != nil && out.Cluster.Identity.Oidc != nil {
				st.OIDCIssuer = aws.ToString(out.Cluster.Identity.Oidc.Issuer)
			}
			log.Info("cluster ACTIVE", "endpoint", st.ClusterEndpoint)
			return st.Save()
		case ekstypes.ClusterStatusFailed:
			return fmt.Errorf("cluster entered FAILED state")
		default:
			log.Info("cluster provisioning", "status", out.Cluster.Status)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Second):
		}
	}
	return fmt.Errorf("cluster did not become ACTIVE within timeout")
}

// CreateNodeGroup provisions a managed node group and waits for ACTIVE.
func CreateNodeGroup(ctx context.Context, log *slog.Logger, c *eks.Client, iamc *iam.Client, st *State) error {
	if st.NodeGroupName == "" {
		st.NodeGroupName = "nuon-" + st.InstallID + "-ng"
	}
	roleArn, err := getRoleARN(ctx, iamc, st.NodeRoleName)
	if err != nil {
		return err
	}
	log = log.With("node_group", st.NodeGroupName)

	_, err = c.CreateNodegroup(ctx, &eks.CreateNodegroupInput{
		ClusterName:   &st.ClusterName,
		NodegroupName: &st.NodeGroupName,
		NodeRole:      &roleArn,
		Subnets:       st.PrivateSubnetIDs,
		ScalingConfig: &ekstypes.NodegroupScalingConfig{
			MinSize:     aws.Int32(1),
			DesiredSize: aws.Int32(2),
			MaxSize:     aws.Int32(3),
		},
		InstanceTypes: []string{"t3.medium"},
	})
	if err != nil {
		var ae smithy.APIError
		if !(errors.As(err, &ae) && ae.ErrorCode() == "ResourceInUseException") {
			return fmt.Errorf("create node group: %w", err)
		}
		log.Info("node group exists, reusing")
	} else {
		log.Info("create node group initiated, waiting for ACTIVE")
	}
	if err := st.Save(); err != nil {
		return err
	}

	deadline := time.Now().Add(30 * time.Minute)
	for time.Now().Before(deadline) {
		out, err := c.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
			ClusterName:   &st.ClusterName,
			NodegroupName: &st.NodeGroupName,
		})
		if err != nil {
			return fmt.Errorf("describe node group: %w", err)
		}
		switch out.Nodegroup.Status {
		case ekstypes.NodegroupStatusActive:
			log.Info("node group ACTIVE")
			return nil
		case ekstypes.NodegroupStatusCreateFailed, ekstypes.NodegroupStatusDegraded:
			return fmt.Errorf("node group entered status %s", out.Nodegroup.Status)
		default:
			log.Info("node group provisioning", "status", out.Nodegroup.Status)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Second):
		}
	}
	return fmt.Errorf("node group did not become ACTIVE within timeout")
}

func getRoleARN(ctx context.Context, c *iam.Client, name string) (string, error) {
	out, err := c.GetRole(ctx, &iam.GetRoleInput{RoleName: &name})
	if err != nil {
		return "", fmt.Errorf("get role %s: %w", name, err)
	}
	return aws.ToString(out.Role.Arn), nil
}

// DeleteEKS tears down node group then cluster. Idempotent.
func DeleteEKS(ctx context.Context, log *slog.Logger, c *eks.Client, st *State) error {
	if st.NodeGroupName != "" && st.ClusterName != "" {
		_, err := c.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
			ClusterName:   &st.ClusterName,
			NodegroupName: &st.NodeGroupName,
		})
		if err != nil {
			var ae smithy.APIError
			if !(errors.As(err, &ae) && ae.ErrorCode() == "ResourceNotFoundException") {
				log.Warn("delete node group", "err", err)
			}
		}
		// Wait for deletion
		deadline := time.Now().Add(30 * time.Minute)
		for time.Now().Before(deadline) {
			_, err := c.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
				ClusterName:   &st.ClusterName,
				NodegroupName: &st.NodeGroupName,
			})
			if err != nil {
				var ae smithy.APIError
				if errors.As(err, &ae) && ae.ErrorCode() == "ResourceNotFoundException" {
					break
				}
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(30 * time.Second):
			}
		}
		st.NodeGroupName = ""
	}
	if st.ClusterName != "" {
		_, err := c.DeleteCluster(ctx, &eks.DeleteClusterInput{Name: &st.ClusterName})
		if err != nil {
			var ae smithy.APIError
			if !(errors.As(err, &ae) && ae.ErrorCode() == "ResourceNotFoundException") {
				log.Warn("delete cluster", "err", err)
			}
		}
		deadline := time.Now().Add(30 * time.Minute)
		for time.Now().Before(deadline) {
			_, err := c.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: &st.ClusterName})
			if err != nil {
				var ae smithy.APIError
				if errors.As(err, &ae) && ae.ErrorCode() == "ResourceNotFoundException" {
					break
				}
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(30 * time.Second):
			}
		}
		st.ClusterName = ""
		st.ClusterEndpoint = ""
		st.ClusterCAData = ""
		st.OIDCIssuer = ""
	}
	return st.Save()
}
