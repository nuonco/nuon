package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
)

type stackPlanAPI interface {
	CreateChangeSet(context.Context, *cloudformation.CreateChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error)
	DescribeChangeSet(context.Context, *cloudformation.DescribeChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error)
	DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	GetTemplate(context.Context, *cloudformation.GetTemplateInput, ...func(*cloudformation.Options)) (*cloudformation.GetTemplateOutput, error)
}

func createStackPlan(ctx context.Context, cfn stackPlanAPI, stackName string, candidate operation.BundleCandidate, candidateTemplate []byte) (*operation.StackCandidate, error) {
	if candidate.Deployment == nil || candidate.Deployment.StackTemplateURL == "" {
		return nil, fmt.Errorf("candidate has no install stack deployment assets")
	}
	stack, err := cfn.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{StackName: &stackName})
	if err != nil {
		return nil, fmt.Errorf("describe install stack parameters: %w", err)
	}
	if len(stack.Stacks) != 1 {
		return nil, fmt.Errorf("describe install stack parameters: expected one stack, got %d", len(stack.Stacks))
	}
	currentTemplate, err := cfn.GetTemplate(ctx, &cloudformation.GetTemplateInput{StackName: &stackName, TemplateStage: cloudformationtypes.TemplateStageOriginal})
	if err != nil {
		return nil, fmt.Errorf("read current install stack template: %w", err)
	}
	parameters := make([]cloudformationtypes.Parameter, 0, len(stack.Stacks[0].Parameters))
	for _, parameter := range stack.Stacks[0].Parameters {
		parameters = append(parameters, cloudformationtypes.Parameter{ParameterKey: parameter.ParameterKey, UsePreviousValue: aws.Bool(true)})
	}
	digest := strings.TrimPrefix(candidate.Bundle.BundleDigest, "sha256:")
	if len(digest) < 12 {
		return nil, fmt.Errorf("candidate bundle digest is invalid")
	}
	changeSetName := fmt.Sprintf("nuon-bundle-%s-%d", digest[:12], time.Now().UTC().UnixMilli())
	created, err := cfn.CreateChangeSet(ctx, &cloudformation.CreateChangeSetInput{
		StackName: &stackName, ChangeSetName: &changeSetName, ChangeSetType: cloudformationtypes.ChangeSetTypeUpdate,
		Description: aws.String("Nuon bundle candidate " + candidate.Bundle.BundleDigest), TemplateURL: &candidate.Deployment.StackTemplateURL,
		Capabilities: []cloudformationtypes.Capability{cloudformationtypes.CapabilityCapabilityNamedIam, cloudformationtypes.CapabilityCapabilityAutoExpand},
		Parameters:   parameters, IncludeNestedStacks: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("create install stack change set: %w", err)
	}
	described, err := waitForStackPlan(ctx, cfn, stackName, aws.ToString(created.Id))
	if err != nil {
		return nil, err
	}
	result := &operation.StackCandidate{
		SchemaVersion: 1, BundleDigest: candidate.Bundle.BundleDigest, StackName: stackName,
		ChangeSetName: changeSetName, ChangeSetARN: aws.ToString(created.Id), TemplateURL: candidate.Deployment.StackTemplateURL,
		CandidateBundleKey: candidate.Deployment.CandidateBundleKey, TargetBundleKey: candidate.Deployment.TargetBundleKey,
		Status: string(described.Status), ExecutionStatus: string(described.ExecutionStatus), StatusReason: aws.ToString(described.StatusReason),
		NoOp: isNoStackPlan(described), Changes: normalizeStackPlanChanges(described.Changes), CreatedAt: time.Now().UTC(),
	}
	for described.NextToken != nil {
		described, err = cfn.DescribeChangeSet(ctx, &cloudformation.DescribeChangeSetInput{ChangeSetName: created.Id, StackName: &stackName, NextToken: described.NextToken})
		if err != nil {
			return nil, fmt.Errorf("describe stack change set changes: %w", err)
		}
		result.Changes = append(result.Changes, normalizeStackPlanChanges(described.Changes)...)
	}
	if err := addStackPropertyChanges(result, []byte(aws.ToString(currentTemplate.TemplateBody)), candidateTemplate); err != nil {
		return nil, fmt.Errorf("compare install stack templates: %w", err)
	}
	return result, nil
}

func waitForStackPlan(ctx context.Context, cfn stackPlanAPI, stackName, changeSetID string) (*cloudformation.DescribeChangeSetOutput, error) {
	for {
		described, err := cfn.DescribeChangeSet(ctx, &cloudformation.DescribeChangeSetInput{ChangeSetName: &changeSetID, StackName: &stackName})
		if err != nil {
			return nil, fmt.Errorf("describe stack change set: %w", err)
		}
		switch described.Status {
		case cloudformationtypes.ChangeSetStatusCreateComplete:
			return described, nil
		case cloudformationtypes.ChangeSetStatusFailed:
			if isNoStackPlan(described) {
				return described, nil
			}
			return nil, fmt.Errorf("create stack change set: %s", aws.ToString(described.StatusReason))
		case cloudformationtypes.ChangeSetStatusDeleteComplete:
			return nil, fmt.Errorf("stack change set was deleted")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func isNoStackPlan(described *cloudformation.DescribeChangeSetOutput) bool {
	if described.Status != cloudformationtypes.ChangeSetStatusFailed {
		return false
	}
	reason := strings.ToLower(aws.ToString(described.StatusReason))
	return strings.Contains(reason, "didn't contain changes") || strings.Contains(reason, "no updates are to be performed")
}

func normalizeStackPlanChanges(changes []cloudformationtypes.Change) []operation.StackChange {
	result := make([]operation.StackChange, 0, len(changes))
	for _, change := range changes {
		if change.ResourceChange == nil {
			continue
		}
		resource := change.ResourceChange
		details := make([]operation.StackChangeDetail, 0, len(resource.Details))
		for _, detail := range resource.Details {
			normalized := operation.StackChangeDetail{Evaluation: string(detail.Evaluation), ChangeSource: string(detail.ChangeSource), CausingEntity: aws.ToString(detail.CausingEntity)}
			if detail.Target != nil {
				normalized.Attribute = string(detail.Target.Attribute)
				normalized.Name = aws.ToString(detail.Target.Name)
				normalized.RequiresRecreation = string(detail.Target.RequiresRecreation)
			}
			details = append(details, normalized)
		}
		scope := make([]string, 0, len(resource.Scope))
		for _, value := range resource.Scope {
			scope = append(scope, string(value))
		}
		result = append(result, operation.StackChange{
			Action: string(resource.Action), LogicalResourceID: aws.ToString(resource.LogicalResourceId), ResourceType: aws.ToString(resource.ResourceType),
			Replacement: string(resource.Replacement), Scope: scope, Details: details,
		})
	}
	return result
}
