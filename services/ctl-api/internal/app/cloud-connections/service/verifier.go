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
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type VerificationResult struct {
	Status       app.CloudConnectionStatus
	Message      string
	Capabilities []app.CloudConnectionCapability
	Registries   []string
}

type Verifier interface {
	Verify(context.Context, *app.CloudConnection, VerifyOptions) (VerificationResult, error)
}

type stsAPI interface {
	AssumeRoleWithWebIdentity(context.Context, *sts.AssumeRoleWithWebIdentityInput, ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error)
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type ecrAPI interface {
	GetAuthorizationToken(context.Context, *ecr.GetAuthorizationTokenInput, ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error)
	DescribeRepositories(context.Context, *ecr.DescribeRepositoriesInput, ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
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
	capabilities := make([]app.CloudConnectionCapability, 0, len(connection.Capabilities))
	for _, capability := range connection.Capabilities {
		switch capability {
		case app.CloudConnectionCapabilityStacks:
			_, err = cloudformation.NewFromConfig(assumed).DescribeStacks(ctx, &cloudformation.DescribeStacksInput{})
		case app.CloudConnectionCapabilityImages:
			ecrClient := ecr.NewFromConfig(assumed)
			_, err = ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
			if err == nil && len(options.Repositories) > 0 {
				_, err = ecrClient.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{RepositoryNames: options.Repositories})
			}
		}
		if err == nil {
			capabilities = append(capabilities, capability)
			continue
		}
		if !isAccessDenied(err) {
			return VerificationResult{}, fmt.Errorf("probe %s capability: %w", capability, err)
		}
	}
	if len(capabilities) == 0 {
		return VerificationResult{Status: app.CloudConnectionStatusError, Message: "The role lacks all requested capabilities.", Capabilities: capabilities}, nil
	}
	return VerificationResult{Status: app.CloudConnectionStatusVerified, Message: "Cloud connection verified.", Capabilities: capabilities}, nil
}

func verificationFailure(message string) VerificationResult {
	return VerificationResult{Status: app.CloudConnectionStatusError, Message: message, Capabilities: []app.CloudConnectionCapability{}}
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
