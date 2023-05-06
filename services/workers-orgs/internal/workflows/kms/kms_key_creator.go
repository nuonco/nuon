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
//go:generate mockgen -destination=mock_kms_key_creator.go -source=kms_key_creator.go -package=kms
type CreateKMSKeyRequest struct {
	AssumeRoleARN string `validate:"required" json:"assume_role_arn"`

	KeyTags [][2]string `validate:"required" json:"key_tags"`
}

func (r CreateKMSKeyRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateKMSKeyResponse struct {
	KeyArn string `validate:"required" json:"key_arn"`
	KeyID  string `validate:"required" json:"key_id"`
}

func (a *Activities) CreateKMSKey(ctx context.Context, req CreateKMSKeyRequest) (CreateKMSKeyResponse, error) {
	var resp CreateKMSKeyResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	assumer, err := assumerole.New(a.validator,
		assumerole.WithRoleARN(req.AssumeRoleARN),
		assumerole.WithRoleSessionName("workers-orgs-kms-key-creator"))
	if err != nil {
		return resp, fmt.Errorf("unable to create role assumer: %w", err)
	}
	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to load config with assumed role: %w", err)
	}

	client := kms.NewFromConfig(cfg)
	keyMeta, err := a.kmsKeyCreator.createKMSKey(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable to create odr IAM policy: %w", err)
	}

	resp.KeyArn = *keyMeta.Arn
	resp.KeyID = *keyMeta.KeyId
	return resp, nil
}

type kmsKeyCreator interface {
	createKMSKey(context.Context, awsClientKMSKeyCreator, CreateKMSKeyRequest) (*kms_types.KeyMetadata, error)
}

var _ kmsKeyCreator = (*kmsKeyCreatorImpl)(nil)

type kmsKeyCreatorImpl struct{}

type awsClientKMSKeyCreator interface {
	CreateKey(context.Context, *kms.CreateKeyInput, ...func(*kms.Options)) (*kms.CreateKeyOutput, error)
}

func (o *kmsKeyCreatorImpl) createKMSKey(ctx context.Context, client awsClientKMSKeyCreator, req CreateKMSKeyRequest) (*kms_types.KeyMetadata, error) {
	tags := make([]kms_types.Tag, 0, len(req.KeyTags)+1)
	for _, pair := range req.KeyTags {
		tags = append(tags, kms_types.Tag{
			TagKey:   generics.ToPtr(pair[0]),
			TagValue: generics.ToPtr(pair[1]),
		})
	}

	params := &kms.CreateKeyInput{
		CustomerMasterKeySpec: kms_types.CustomerMasterKeySpecSymmetricDefault,
		KeyUsage:              kms_types.KeyUsageTypeEncryptDecrypt,
		Origin:                kms_types.OriginTypeAwsKms,
		Tags:                  tags,
	}
	output, err := client.CreateKey(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to create policy: %w", err)
	}
	return output.KeyMetadata, nil
}
