package service

import (
	"context"
	"errors"
	"testing"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go"
)

type fakeSTS struct {
	err               error
	external          *string
	assumeRole        func(*sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error)
	getCallerIdentity func() (*sts.GetCallerIdentityOutput, error)
}

func (f *fakeSTS) AssumeRole(_ context.Context, input *sts.AssumeRoleInput, _ ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	f.external = input.ExternalId
	if f.assumeRole != nil {
		return f.assumeRole(input)
	}
	return &sts.AssumeRoleOutput{}, f.err
}

func (f *fakeSTS) GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	if f.getCallerIdentity != nil {
		return f.getCallerIdentity()
	}
	return nil, errors.New("unused")
}

func TestSucceedsAssumeProbesExternalID(t *testing.T) {
	fake := &fakeSTS{}
	succeeded, err := probeAssume(context.Background(), fake, "arn:aws:iam::123456789012:role/test", "session", "wrong")
	if err != nil || !succeeded {
		t.Fatal("expected successful probe")
	}
	if fake.external == nil || *fake.external != "wrong" {
		t.Fatal("expected wrong external ID on probe")
	}
	fake.err = &smithy.GenericAPIError{Code: "AccessDenied", Message: "denied"}
	succeeded, err = probeAssume(context.Background(), fake, "arn:aws:iam::123456789012:role/test", "session", "")
	if err != nil || succeeded {
		t.Fatal("expected failed probe")
	}
	if fake.external != nil {
		t.Fatal("expected no external ID on negative probe")
	}
	fake.err = errors.New("network")
	if _, err := probeAssume(context.Background(), fake, "arn:aws:iam::123456789012:role/test", "session", ""); err == nil {
		t.Fatal("expected runtime probe error")
	}
}

func TestVerifyConnection(t *testing.T) {
	accessDenied := &smithy.GenericAPIError{Code: "AccessDenied", Message: "denied"}
	credentials := &sts_types.Credentials{
		AccessKeyId:     aws.String("access-key"),
		SecretAccessKey: aws.String("secret-key"),
		SessionToken:    aws.String("session-token"),
	}
	baseSTS := &fakeSTS{assumeRole: func(input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
		if *input.RoleArn != "arn:aws:iam::999999999999:role/management" || input.ExternalId != nil {
			t.Fatalf("unexpected management role assumption: %#v", input)
		}
		return &sts.AssumeRoleOutput{Credentials: credentials}, nil
	}}
	managementSTS := &fakeSTS{assumeRole: func(input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
		if *input.RoleArn != "arn:aws:iam::123456789012:role/nuon" || *input.RoleSessionName != "nuon-connection-verification" {
			t.Fatalf("customer role probes must use the same role and session: %#v", input)
		}
		if input.ExternalId == nil || *input.ExternalId != "correct-external-id" {
			return nil, accessDenied
		}
		return &sts.AssumeRoleOutput{Credentials: credentials, AssumedRoleUser: &sts_types.AssumedRoleUser{
			Arn:           aws.String("arn:aws:sts::123456789012:assumed-role/nuon/nuon-connection-verification"),
			AssumedRoleId: aws.String("role-id:nuon-connection-verification"),
		}}, nil
	}}
	identitySTS := &fakeSTS{getCallerIdentity: func() (*sts.GetCallerIdentityOutput, error) {
		return &sts.GetCallerIdentityOutput{
			Account: aws.String("123456789012"),
			Arn:     aws.String("arn:aws:sts::123456789012:assumed-role/nuon/nuon-connection-verification"),
			UserId:  aws.String("role-id:nuon-connection-verification"),
		}, nil
	}}
	stsClients := []stsAPI{baseSTS, managementSTS, identitySTS}
	verifier := &awsVerifier{
		loadBase: func(context.Context, string) (aws.Config, error) { return aws.Config{}, nil },
		newSTS: func(aws.Config) stsAPI {
			client := stsClients[0]
			stsClients = stsClients[1:]
			return client
		},
		randomID: func() (string, error) { return "wrong-external-id", nil },
	}

	result, err := verifier.Verify(context.Background(), "arn:aws:iam::999999999999:role/management", "arn:aws:iam::123456789012:role/nuon", "correct-external-id", "123456789012")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "verified" || result.PrincipalARN == "" {
		t.Fatalf("unexpected verification result: %#v", result)
	}
}

func TestMatchesAssumedRoleARN(t *testing.T) {
	roleARN := "arn:aws:iam::123456789012:role/path/nuon"
	if !matchesAssumedRoleARN("arn:aws:sts::123456789012:assumed-role/nuon/session", roleARN, "session") {
		t.Fatal("expected exact assumed role ARN to match")
	}
	for _, value := range []string{
		"arn:aws:sts::123456789012:assumed-role/other/session",
		"arn:aws:sts::123456789012:assumed-role/nuon/other-session",
		"arn:aws:sts::999999999999:assumed-role/nuon/session",
	} {
		if matchesAssumedRoleARN(value, roleARN, "session") {
			t.Fatalf("unexpected match for %s", value)
		}
	}
}

func TestVerifyRejectsMismatchedPrincipal(t *testing.T) {
	const expectedARN = "arn:aws:sts::123456789012:assumed-role/nuon/nuon-connection-verification"
	const expectedID = "role-id:nuon-connection-verification"
	for _, test := range []struct {
		name   string
		arn    string
		userID string
	}{
		{name: "different role", arn: "arn:aws:sts::123456789012:assumed-role/other/nuon-connection-verification", userID: expectedID},
		{name: "different ARN", arn: "arn:aws:sts::123456789012:assumed-role/nuon/other-session", userID: expectedID},
		{name: "different user ID", arn: expectedARN, userID: "other-role-id:nuon-connection-verification"},
	} {
		t.Run(test.name, func(t *testing.T) {
			credentials := &sts_types.Credentials{AccessKeyId: aws.String("access-key"), SecretAccessKey: aws.String("secret-key"), SessionToken: aws.String("session-token")}
			accessDenied := &smithy.GenericAPIError{Code: "AccessDenied", Message: "denied"}
			baseSTS := &fakeSTS{assumeRole: func(*sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
				return &sts.AssumeRoleOutput{Credentials: credentials}, nil
			}}
			managementSTS := &fakeSTS{assumeRole: func(input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
				if input.ExternalId == nil || *input.ExternalId != "correct-external-id" {
					return nil, accessDenied
				}
				return &sts.AssumeRoleOutput{Credentials: credentials, AssumedRoleUser: &sts_types.AssumedRoleUser{Arn: aws.String(expectedARN), AssumedRoleId: aws.String(expectedID)}}, nil
			}}
			identitySTS := &fakeSTS{getCallerIdentity: func() (*sts.GetCallerIdentityOutput, error) {
				return &sts.GetCallerIdentityOutput{Account: aws.String("123456789012"), Arn: aws.String(test.arn), UserId: aws.String(test.userID)}, nil
			}}
			stsClients := []stsAPI{baseSTS, managementSTS, identitySTS}
			verifier := &awsVerifier{
				loadBase: func(context.Context, string) (aws.Config, error) { return aws.Config{}, nil },
				newSTS: func(aws.Config) stsAPI {
					client := stsClients[0]
					stsClients = stsClients[1:]
					return client
				},
				randomID: func() (string, error) { return "wrong-external-id", nil },
			}

			result, err := verifier.Verify(context.Background(), "arn:aws:iam::999999999999:role/management", "arn:aws:iam::123456789012:role/nuon", "correct-external-id", "123456789012")
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != "error" || result.Code != "principal_mismatch" {
				t.Fatalf("unexpected verification result: %#v", result)
			}
		})
	}
}

func TestVerifyRejectsRoleWithoutExternalIDRequirement(t *testing.T) {
	credentials := &sts_types.Credentials{
		AccessKeyId:     aws.String("access-key"),
		SecretAccessKey: aws.String("secret-key"),
		SessionToken:    aws.String("session-token"),
	}
	baseSTS := &fakeSTS{assumeRole: func(*sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
		return &sts.AssumeRoleOutput{Credentials: credentials}, nil
	}}
	managementSTS := &fakeSTS{assumeRole: func(*sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
		return &sts.AssumeRoleOutput{Credentials: credentials}, nil
	}}
	stsClients := []stsAPI{baseSTS, managementSTS}
	verifier := &awsVerifier{
		loadBase: func(context.Context, string) (aws.Config, error) { return aws.Config{}, nil },
		newSTS: func(aws.Config) stsAPI {
			client := stsClients[0]
			stsClients = stsClients[1:]
			return client
		},
		randomID: func() (string, error) { return "wrong-external-id", nil },
	}

	result, err := verifier.Verify(context.Background(), "arn:aws:iam::999999999999:role/management", "arn:aws:iam::123456789012:role/nuon", "correct-external-id", "123456789012")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "error" || result.Code != "external_id_not_required" {
		t.Fatalf("unexpected verification result: %#v", result)
	}
}
