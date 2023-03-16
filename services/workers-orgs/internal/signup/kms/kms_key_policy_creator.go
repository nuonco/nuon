package kms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws-assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_kms_key_policy_creator.go -source=kms_key_policy_creator.go -package=kms
type CreateKMSKeyPolicyRequest struct {
	AssumeRoleARN string `validate:"required"`

	KeyID      string `validate:"required"`
	PolicyName string `validate:"required"`
	Policy     string `validate:"required"`
}

type CreateKMSKeyPolicyResponse struct{}

func (a *Activities) CreateKMSKeyPolicy(ctx context.Context, req CreateKMSKeyPolicyRequest) (CreateKMSKeyPolicyResponse, error) {
	var resp CreateKMSKeyPolicyResponse
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
	err = a.kmsKeyPolicyCreator.createKMSKeyPolicy(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable to create odr IAM role: %w", err)
	}

	return resp, nil
}

func (r CreateKMSKeyPolicyRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type kmsKeyPolicyCreator interface {
	createKMSKeyPolicy(context.Context, awsClientKMSKeyPolicyCreator, CreateKMSKeyPolicyRequest) error
}

var _ kmsKeyPolicyCreator = (*kmsKeyPolicyCreatorImpl)(nil)

type kmsKeyPolicyCreatorImpl struct{}

type awsClientKMSKeyPolicyCreator interface {
	PutKeyPolicy(context.Context, *kms.PutKeyPolicyInput, ...func(*kms.Options)) (*kms.PutKeyPolicyOutput, error)
}

func (o *kmsKeyPolicyCreatorImpl) createKMSKeyPolicy(ctx context.Context, client awsClientKMSKeyPolicyCreator, req CreateKMSKeyPolicyRequest) error {
	params := &kms.PutKeyPolicyInput{
		KeyId:      generics.ToPtr(req.KeyID),
		Policy:     generics.ToPtr(req.Policy),
		PolicyName: generics.ToPtr(req.PolicyName),
	}
	_, err := client.PutKeyPolicy(ctx, params)
	if err != nil {
		return err
	}

	return nil
}
