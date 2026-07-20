package iam

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/go-playground/validator/v10"
)

type recordingSTS struct {
	assumeRoleInput *sts.AssumeRoleInput
}

func (r *recordingSTS) AssumeRole(_ context.Context, input *sts.AssumeRoleInput, _ ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	r.assumeRoleInput = input
	return &sts.AssumeRoleOutput{Credentials: &sts_types.Credentials{}}, nil
}

func (r *recordingSTS) AssumeRoleWithWebIdentity(_ context.Context, _ *sts.AssumeRoleWithWebIdentityInput, _ ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	return &sts.AssumeRoleWithWebIdentityOutput{Credentials: &sts_types.Credentials{}}, nil
}

func TestExternalIDRejectsOIDC(t *testing.T) {
	_, err := New(validator.New(), WithSettings(Settings{RoleARN: "arn:aws:iam::123456789012:role/test", RoleSessionName: "test", ExternalID: "external", UseGithubOIDC: true}))
	if err == nil {
		t.Fatal("expected external ID with OIDC to be rejected")
	}
}

func TestAssumeRoleExternalID(t *testing.T) {
	assumer, err := New(validator.New(), WithSettings(Settings{
		RoleARN:         "arn:aws:iam::123456789012:role/test",
		RoleSessionName: "test",
		ExternalID:      "connection-external-id",
	}))
	if err != nil {
		t.Fatal(err)
	}

	client := &recordingSTS{}
	if _, err := assumer.assumeIamRole(context.Background(), client, assumer.RoleARN, assumer.ExternalID); err != nil {
		t.Fatal(err)
	}
	if client.assumeRoleInput.ExternalId == nil || *client.assumeRoleInput.ExternalId != assumer.ExternalID {
		t.Fatalf("expected final role external ID, got %#v", client.assumeRoleInput.ExternalId)
	}

	if _, err := assumer.assumeIamRole(context.Background(), client, "arn:aws:iam::123456789012:role/broker", ""); err != nil {
		t.Fatal(err)
	}
	if client.assumeRoleInput.ExternalId != nil {
		t.Fatalf("intermediate role received external ID: %q", *client.assumeRoleInput.ExternalId)
	}
}
