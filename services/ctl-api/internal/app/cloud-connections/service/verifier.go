package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type VerificationResult struct {
	Status  app.CloudConnectionStatus
	Message string
}

type VerifyOptions struct {
	IdentityOnly bool
}

type Verifier interface {
	Verify(context.Context, *app.CloudConnection, VerifyOptions) (VerificationResult, error)
}

type stsAPI interface {
	AssumeRoleWithWebIdentity(context.Context, *sts.AssumeRoleWithWebIdentityInput, ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error)
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type cloudFormationAPI interface {
	DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
}

type awsVerifier struct {
	issuer *oidcissuer.Issuer
}

func NewAWSVerifier(issuer *oidcissuer.Issuer) Verifier {
	return &awsVerifier{issuer: issuer}
}

func (v *awsVerifier) Verify(ctx context.Context, connection *app.CloudConnection, options VerifyOptions) (VerificationResult, error) {
	if connection.Platform != app.CloudPlatformAWS {
		return VerificationResult{}, fmt.Errorf("only aws cloud connections are supported")
	}
	if v.issuer == nil {
		return VerificationResult{}, fmt.Errorf("cloud connection OIDC issuer is unavailable")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	base, err := config.LoadDefaultConfig(ctx, config.WithRegion(connection.DefaultRegion))
	if err != nil {
		return VerificationResult{}, fmt.Errorf("load AWS config: %w", err)
	}
	client := sts.NewFromConfig(base)
	token, err := v.issuer.Mint(ctx, subject(connection), "sts.amazonaws.com", 10*time.Minute)
	if err != nil {
		return VerificationResult{}, err
	}
	output, err := client.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{RoleArn: &connection.Principal, RoleSessionName: aws.String("nuon-cloud-connection-verification"), WebIdentityToken: &token, DurationSeconds: aws.Int32(900)})
	if err != nil {
		if isAccessDenied(err) {
			return verificationFailure("Nuon OIDC identity is not trusted by this role."), nil
		}
		return VerificationResult{}, err
	}
	foreignToken, err := v.issuer.Mint(ctx, "org:foreign:connection:foreign", "sts.amazonaws.com", 10*time.Minute)
	if err != nil {
		return VerificationResult{}, err
	}
	if _, err := client.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{RoleArn: &connection.Principal, RoleSessionName: aws.String("nuon-cloud-connection-negative-probe"), WebIdentityToken: &foreignToken, DurationSeconds: aws.Int32(900)}); err == nil {
		return verificationFailure("The role trust policy accepts a foreign Nuon connection subject."), nil
	} else if !isAccessDenied(err) {
		return VerificationResult{}, fmt.Errorf("probe foreign subject: %w", err)
	}
	assumed, err := awsConfigWithCredentials(base, output.Credentials)
	if err != nil {
		return VerificationResult{}, err
	}
	identity, err := sts.NewFromConfig(assumed).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return verificationFailure("The assumed role identity could not be verified."), nil
	}
	if identity.Account == nil || *identity.Account != connection.TargetID || identity.Arn == nil || !matchesRole(*identity.Arn, connection.Principal) {
		return verificationFailure("The assumed identity does not match the configured target and principal."), nil
	}
	if connection.Preset == app.CloudConnectionPresetStacks && !options.IdentityOnly {
		if _, err := cloudformation.NewFromConfig(assumed).DescribeStacks(ctx, &cloudformation.DescribeStacksInput{}); err != nil {
			if isAccessDenied(err) {
				return verificationFailure("The role lacks CloudFormation read access required by the stacks preset."), nil
			}
			return VerificationResult{}, fmt.Errorf("probe CloudFormation access: %w", err)
		}
	}
	return VerificationResult{Status: app.CloudConnectionStatusVerified, Message: "Cloud connection verified."}, nil
}

func verificationFailure(message string) VerificationResult {
	return VerificationResult{Status: app.CloudConnectionStatusError, Message: message}
}

func awsConfigWithCredentials(base aws.Config, value *stsTypes.Credentials) (aws.Config, error) {
	if value == nil || value.AccessKeyId == nil || value.SecretAccessKey == nil || value.SessionToken == nil {
		return aws.Config{}, fmt.Errorf("assume role response did not include credentials")
	}
	result := base.Copy()
	result.Credentials = credentials.NewStaticCredentialsProvider(*value.AccessKeyId, *value.SecretAccessKey, *value.SessionToken)
	return result, nil
}

func matchesRole(actual, requested string) bool {
	actualARN, err := awsarn.Parse(actual)
	if err != nil {
		return false
	}
	requestedARN, err := awsarn.Parse(requested)
	if err != nil {
		return false
	}
	roleName := requestedARN.Resource[strings.LastIndex(requestedARN.Resource, "/")+1:]
	return actualARN.Partition == requestedARN.Partition && actualARN.Service == "sts" && actualARN.AccountID == requestedARN.AccountID && strings.HasPrefix(actualARN.Resource, "assumed-role/"+roleName+"/")
}

func isAccessDenied(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "AccessDenied" || apiErr.ErrorCode() == "AccessDeniedException")
}
