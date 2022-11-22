package runner

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
)

type iamRoleAssumer interface {
	assumeIamRole(context.Context, string, awsClientIamRoleAssumer) (*sts_types.Credentials, error)
}

type iamRoleAssumerImpl struct{}

var _ iamRoleAssumer = (*iamRoleAssumerImpl)(nil)

type awsClientIamRoleAssumer interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

func (r *iamRoleAssumerImpl) assumeIamRole(ctx context.Context, roleArn string, client awsClientIamRoleAssumer) (*sts_types.Credentials, error) {
	params := &sts.AssumeRoleInput{
		RoleArn:         toPtr(roleArn),
		RoleSessionName: toPtr("workers-orgs"),
	}
	resp, err := client.AssumeRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Credentials, nil
}
