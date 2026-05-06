package stack

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
)

const (
	eksClusterTrust = `{
  "Version": "2012-10-17",
  "Statement": [{"Effect":"Allow","Principal":{"Service":"eks.amazonaws.com"},"Action":"sts:AssumeRole"}]
}`
	eksNodeTrust = `{
  "Version": "2012-10-17",
  "Statement": [{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action":"sts:AssumeRole"}]
}`

	clusterPolicyARN     = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
	nodeWorkerPolicyARN  = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
	nodeCNIPolicyARN     = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
	nodeECRReadPolicyARN = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
)

func CreateIAMRoles(ctx context.Context, log *slog.Logger, c *iam.Client, st *State) error {
	if st.ClusterRoleName == "" {
		st.ClusterRoleName = "nuon-" + st.InstallID + "-eks-cluster"
		if err := createRoleWithPolicies(ctx, c, st.ClusterRoleName, eksClusterTrust, []string{clusterPolicyARN}, st.InstallID); err != nil {
			return err
		}
		log.Info("created EKS cluster role", "role", st.ClusterRoleName)
		if err := st.Save(); err != nil {
			return err
		}
	}
	if st.NodeRoleName == "" {
		st.NodeRoleName = "nuon-" + st.InstallID + "-eks-node"
		if err := createRoleWithPolicies(ctx, c, st.NodeRoleName, eksNodeTrust,
			[]string{nodeWorkerPolicyARN, nodeCNIPolicyARN, nodeECRReadPolicyARN}, st.InstallID); err != nil {
			return err
		}
		log.Info("created EKS node role", "role", st.NodeRoleName)
		if err := st.Save(); err != nil {
			return err
		}
	}
	return nil
}

func createRoleWithPolicies(ctx context.Context, c *iam.Client, name, trust string, policies []string, installID string) error {
	_, err := c.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 &name,
		AssumeRolePolicyDocument: &trust,
		Tags: []iamtypes.Tag{
			{Key: aws.String("nuon:install_id"), Value: &installID},
		},
	})
	if err != nil {
		var ae smithy.APIError
		if !(errors.As(err, &ae) && ae.ErrorCode() == "EntityAlreadyExists") {
			return fmt.Errorf("create role %s: %w", name, err)
		}
	}
	for _, p := range policies {
		p := p
		if _, err := c.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{RoleName: &name, PolicyArn: &p}); err != nil {
			return fmt.Errorf("attach %s -> %s: %w", p, name, err)
		}
	}
	return nil
}

func DeleteIAMRoles(ctx context.Context, log *slog.Logger, c *iam.Client, st *State) error {
	for _, name := range []*string{&st.ClusterRoleName, &st.NodeRoleName} {
		if *name == "" {
			continue
		}
		out, err := c.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: name})
		if err == nil {
			for _, p := range out.AttachedPolicies {
				_, _ = c.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{RoleName: name, PolicyArn: p.PolicyArn})
			}
		}
		if _, err := c.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: name}); err != nil {
			log.Warn("delete role", "role", *name, "err", err)
		} else {
			log.Info("deleted role", "role", *name)
		}
		*name = ""
	}
	return st.Save()
}
