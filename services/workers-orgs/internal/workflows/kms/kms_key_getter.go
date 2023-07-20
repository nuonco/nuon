package kms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	kms_types "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_kms_key_getter.go -source=kms_key_getter.go -package=kms
type GetKMSKeyRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	KeyARN string `validate:"required" json:"key_arn"`
}

func (r GetKMSKeyRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type GetKMSKeyResponse struct {
	KeyArn string `validate:"required" json:"key_arn"`
	KeyID  string `validate:"required" json:"key_id"`
}

func (a *Activities) GetKMSKey(ctx context.Context, req *GetKMSKeyRequest) (*GetKMSKeyResponse, error) {
	var resp GetKMSKeyResponse
	if err := req.validate(); err != nil {
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	assumer, err := assumerole.New(a.validator,
		assumerole.WithRoleARN(req.AssumeRoleARN),
		assumerole.WithRoleSessionName("workers-orgs-kms-key-getter"))
	if err != nil {
		return nil, fmt.Errorf("unable to create role assumer: %w", err)
	}
	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load config with assumed role: %w", err)
	}

	client := kms.NewFromConfig(cfg)
	keyMeta, err := a.kmsKeyGetter.getKMSKey(ctx, client, req)
	if err != nil {
		return nil, fmt.Errorf("unable to create odr IAM policy: %w", err)
	}

	resp.KeyArn = *keyMeta.Arn
	resp.KeyID = *keyMeta.KeyId
	return &resp, nil
}

type kmsKeyGetter interface {
	getKMSKey(context.Context, awsClientKMSKeyGetter, *GetKMSKeyRequest) (*kms_types.KeyMetadata, error)
}

var _ kmsKeyGetter = (*kmsKeyGetterImpl)(nil)

type kmsKeyGetterImpl struct{}

type awsClientKMSKeyGetter interface {
	DescribeKey(context.Context, *kms.DescribeKeyInput, ...func(*kms.Options)) (*kms.DescribeKeyOutput, error)
}

func (o *kmsKeyGetterImpl) getKMSKey(ctx context.Context, client awsClientKMSKeyGetter, req *GetKMSKeyRequest) (*kms_types.KeyMetadata, error) {
	params := &kms.DescribeKeyInput{
		KeyId: generics.ToPtr(req.KeyARN),
	}
	output, err := client.DescribeKey(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to get kms key: %w", err)
	}
	return output.KeyMetadata, nil
}
