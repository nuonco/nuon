package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go"
)

type VerificationResult struct {
	Status       string
	Code         string
	Message      string
	PrincipalARN string
}

type Verifier interface {
	Verify(context.Context, string, string, string, string) (VerificationResult, error)
}

type stsAPI interface {
	AssumeRole(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type awsVerifier struct {
	loadBase func(context.Context, string) (aws.Config, error)
	newSTS   func(aws.Config) stsAPI
	randomID func() (string, error)
}

func NewAWSVerifier() Verifier {
	return &awsVerifier{
		loadBase: func(ctx context.Context, region string) (aws.Config, error) {
			return config.LoadDefaultConfig(ctx, config.WithRegion(region))
		},
		newSTS:   func(cfg aws.Config) stsAPI { return sts.NewFromConfig(cfg) },
		randomID: randomExternalID,
	}
}

func (v *awsVerifier) Verify(ctx context.Context, managementRoleARN, roleARN, externalID, expectedAccount string) (VerificationResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	base, err := v.loadBase(ctx, "us-east-1")
	if err != nil {
		return VerificationResult{}, fmt.Errorf("unable to load AWS credentials: %w", err)
	}
	managementOutput, err := assumeRole(ctx, v.newSTS(base), managementRoleARN, "nuon-connection-management", "")
	if err != nil {
		return VerificationResult{}, fmt.Errorf("unable to assume configured management role: %w", err)
	}
	managementConfig, err := configWithCredentials(base, managementOutput.Credentials)
	if err != nil {
		return VerificationResult{}, err
	}
	managementSTS := v.newSTS(managementConfig)
	const sessionName = "nuon-connection-verification"
	withoutExternalID, err := probeAssume(ctx, managementSTS, roleARN, sessionName, "")
	if err != nil {
		return VerificationResult{}, err
	}
	wrongExternalID, err := v.randomID()
	if err != nil {
		return VerificationResult{}, fmt.Errorf("unable to create trust probe external ID: %w", err)
	}
	withWrongExternalID, err := probeAssume(ctx, managementSTS, roleARN, sessionName, wrongExternalID)
	if err != nil {
		return VerificationResult{}, err
	}
	if withoutExternalID || withWrongExternalID {
		return verificationError("external_id_not_required", "The role trust policy must require the provided external ID."), nil
	}
	assumedOutput, err := assumeRole(ctx, managementSTS, roleARN, sessionName, externalID)
	if err != nil {
		if isAccessDenied(err) {
			return verificationError("access_denied", "The management principal could not assume the role with the provided external ID."), nil
		}
		return VerificationResult{}, fmt.Errorf("management role assumption failed: %w", err)
	}
	assumedConfig, err := configWithCredentials(base, assumedOutput.Credentials)
	if err != nil {
		return VerificationResult{}, err
	}
	identity, err := v.newSTS(assumedConfig).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return verificationError("identity_failed", "The assumed role identity could not be verified."), nil
	}
	if identity.Account == nil || *identity.Account != expectedAccount {
		return verificationError("account_mismatch", "The assumed role belongs to a different AWS account."), nil
	}
	if assumedOutput.AssumedRoleUser == nil || assumedOutput.AssumedRoleUser.Arn == nil || assumedOutput.AssumedRoleUser.AssumedRoleId == nil || identity.Arn == nil || identity.UserId == nil || *identity.Arn != *assumedOutput.AssumedRoleUser.Arn || *identity.UserId != *assumedOutput.AssumedRoleUser.AssumedRoleId || !matchesAssumedRoleARN(*identity.Arn, roleARN, sessionName) {
		return verificationError("principal_mismatch", "The verified identity does not match the requested assumed role."), nil
	}
	return VerificationResult{Status: "verified", PrincipalARN: *identity.Arn, Message: "AWS account connection verified."}, nil
}

func probeAssume(ctx context.Context, client stsAPI, roleARN, sessionName, externalID string) (bool, error) {
	_, err := assumeRole(ctx, client, roleARN, sessionName, externalID)
	if err == nil {
		return true, nil
	}
	if isAccessDenied(err) {
		return false, nil
	}
	return false, fmt.Errorf("unable to probe role trust: %w", err)
}

func matchesAssumedRoleARN(value, roleARN, sessionName string) bool {
	requested, err := awsarn.Parse(roleARN)
	if err != nil {
		return false
	}
	actual, err := awsarn.Parse(value)
	if err != nil {
		return false
	}
	roleName := requested.Resource[strings.LastIndex(requested.Resource, "/")+1:]
	return actual.Partition == requested.Partition && actual.Service == "sts" && actual.Region == "" && actual.AccountID == requested.AccountID && actual.Resource == "assumed-role/"+roleName+"/"+sessionName
}

func assumeRole(ctx context.Context, client stsAPI, roleARN, sessionName, externalID string) (*sts.AssumeRoleOutput, error) {
	input := &sts.AssumeRoleInput{RoleArn: &roleARN, RoleSessionName: &sessionName, DurationSeconds: aws.Int32(900)}
	if externalID != "" {
		input.ExternalId = &externalID
	}
	return client.AssumeRole(ctx, input)
}

func configWithCredentials(base aws.Config, value *sts_types.Credentials) (aws.Config, error) {
	if value == nil || value.AccessKeyId == nil || value.SecretAccessKey == nil || value.SessionToken == nil {
		return aws.Config{}, fmt.Errorf("assume role response did not include credentials")
	}
	result := base.Copy()
	result.Credentials = credentials.NewStaticCredentialsProvider(*value.AccessKeyId, *value.SecretAccessKey, *value.SessionToken)
	return result, nil
}

func isAccessDenied(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "AccessDenied" || apiErr.ErrorCode() == "AccessDeniedException")
}

func randomExternalID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func verificationError(code, message string) VerificationResult {
	return VerificationResult{Status: "error", Code: code, Message: message}
}
