package kms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_kms_key_alias_creator.go -source=kms_key_alias_creator.go -package=kms
type CreateKMSKeyAliasRequest struct {
	AssumeRoleARN string `validate:"required"`

	KeyID string `validate:"required"`
	Alias string `validate:"required"`
}

type CreateKMSKeyAliasResponse struct{}

func (a *Activities) CreateKMSKeyAlias(ctx context.Context, req CreateKMSKeyAliasRequest) (CreateKMSKeyAliasResponse, error) {
	var resp CreateKMSKeyAliasResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	assumer, err := assumerole.New(a.validator, assumerole.WithRoleARN(req.AssumeRoleARN), assumerole.WithRoleSessionName("workers-orgs-iam-policy-creator"))
	if err != nil {
		return resp, fmt.Errorf("unable to create role assumer: %w", err)
	}
	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to load config with assumed role: %w", err)
	}

	client := kms.NewFromConfig(cfg)
	err = a.kmsKeyAliasCreator.createKMSKeyAlias(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable to create kms key policy: %w", err)
	}

	return resp, nil
}

func (r CreateKMSKeyAliasRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type kmsKeyAliasCreator interface {
	createKMSKeyAlias(context.Context, awsClientKMSKeyAliasCreator, CreateKMSKeyAliasRequest) error
}

var _ kmsKeyAliasCreator = (*kmsKeyAliasCreatorImpl)(nil)

type kmsKeyAliasCreatorImpl struct{}

type awsClientKMSKeyAliasCreator interface {
	CreateAlias(context.Context, *kms.CreateAliasInput, ...func(*kms.Options)) (*kms.CreateAliasOutput, error)
}

func (o *kmsKeyAliasCreatorImpl) createKMSKeyAlias(ctx context.Context, client awsClientKMSKeyAliasCreator, req CreateKMSKeyAliasRequest) error {
	params := &kms.CreateAliasInput{
		AliasName:   generics.ToPtr(req.Alias),
		TargetKeyId: generics.ToPtr(req.KeyID),
	}
	_, err := client.CreateAlias(ctx, params)
	if err != nil {
		return err
	}

	return nil
}
