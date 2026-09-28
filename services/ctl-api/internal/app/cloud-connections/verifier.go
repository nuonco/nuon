package cloudconnections

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
	IdentityOnly        bool
	RetryIAMPropagation bool
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
	sleep  func(context.Context, time.Duration) error
	now    func() time.Time
}

func NewAWSVerifier(issuer *oidcissuer.Issuer) Verifier {
	return &awsVerifier{issuer: issuer, sleep: sleepWithContext, now: time.Now}
}

func (v *awsVerifier) Verify(ctx context.Context, connection *app.CloudConnection, options VerifyOptions) (VerificationResult, error) {
	if connection.Platform != app.CloudPlatformAWS {
		return VerificationResult{}, fmt.Errorf("only aws cloud connections are supported")
	}
	if v.issuer == nil {
		return VerificationResult{}, fmt.Errorf("cloud connection OIDC issuer is unavailable")
	}
	timeout := 30 * time.Second
	if options.RetryIAMPropagation {
		timeout = 80 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	base, err := config.LoadDefaultConfig(ctx, config.WithRegion(connection.DefaultRegion), config.WithRetryMaxAttempts(1))
	if err != nil {
		return VerificationResult{}, fmt.Errorf("load AWS config: %w", err)
	}
	client := sts.NewFromConfig(base)
	token, err := v.issuer.Mint(ctx, fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID), "sts.amazonaws.com", 10*time.Minute)
	if err != nil {
		return VerificationResult{}, err
	}
	output, err := v.assumeRole(ctx, client, &sts.AssumeRoleWithWebIdentityInput{RoleArn: &connection.Principal, RoleSessionName: aws.String("nuon-cloud-connection-verification"), WebIdentityToken: &token, DurationSeconds: aws.Int32(900)}, options.RetryIAMPropagation)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			return verificationFailure(AssumeRoleErrorMessage(err)), nil
		}
		return VerificationResult{}, err
	}
	foreignToken, err := v.issuer.Mint(ctx, "org:foreign:connection:foreign", "sts.amazonaws.com", 10*time.Minute)
	if err != nil {
		return VerificationResult{}, err
	}
	if _, err := client.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{RoleArn: &connection.Principal, RoleSessionName: aws.String("nuon-cloud-connection-negative-probe"), WebIdentityToken: &foreignToken, DurationSeconds: aws.Int32(900)}); err == nil {
		return verificationFailure("The role trust policy accepts a foreign Nuon connection subject."), nil
	} else if !IsAccessDenied(err) {
		return VerificationResult{}, fmt.Errorf("probe foreign subject: %w", err)
	}
	assumed, err := awsConfigWithCredentials(base, output.Credentials)
	if err != nil {
		return VerificationResult{}, err
	}
	identity, err := sts.NewFromConfig(assumed).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return VerificationResult{}, fmt.Errorf("verify assumed role identity: %w", err)
	}
	if identity.Account == nil || *identity.Account != connection.TargetID || identity.Arn == nil || !matchesRole(*identity.Arn, connection.Principal) {
		return verificationFailure("The assumed identity does not match the configured target and principal."), nil
	}
	if connection.Preset == app.CloudConnectionPresetStacks && !options.IdentityOnly {
		if _, err := cloudformation.NewFromConfig(assumed).DescribeStacks(ctx, &cloudformation.DescribeStacksInput{}); err != nil {
			if IsAccessDenied(err) {
				return verificationFailure("The role lacks CloudFormation read access required to manage install stacks."), nil
			}
			return VerificationResult{}, fmt.Errorf("probe CloudFormation access: %w", err)
		}
	}
	return VerificationResult{Status: app.CloudConnectionStatusVerified, Message: "Cloud connection verified."}, nil
}

func verificationFailure(message string) VerificationResult {
	return VerificationResult{Status: app.CloudConnectionStatusError, Message: message}
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (v *awsVerifier) assumeRole(ctx context.Context, client stsAPI, input *sts.AssumeRoleWithWebIdentityInput, retry bool) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	if !retry {
		return client.AssumeRoleWithWebIdentity(ctx, input)
	}
	retryCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	deadline := v.now().Add(time.Minute)
	var lastDenied error
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if lastDenied != nil && (retryCtx.Err() != nil || !v.now().Before(deadline)) {
			return nil, lastDenied
		}
		output, err := client.AssumeRoleWithWebIdentity(retryCtx, input)
		if !IsAccessDenied(err) {
			if lastDenied != nil && retryCtx.Err() != nil && ctx.Err() == nil {
				return nil, lastDenied
			}
			return output, err
		}
		lastDenied = err
		if err := v.sleep(retryCtx, min(5*time.Second, max(0, deadline.Sub(v.now())))); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, lastDenied
		}
	}
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
