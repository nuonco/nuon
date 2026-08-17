package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
)

const maxInstallStackEvents = 500

type installStackReader interface {
	Read(context.Context, string) (*installStackStatus, error)
}

type cloudFormationClient interface {
	DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	DescribeStackEvents(context.Context, *cloudformation.DescribeStackEventsInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error)
	GetTemplate(context.Context, *cloudformation.GetTemplateInput, ...func(*cloudformation.Options)) (*cloudformation.GetTemplateOutput, error)
}

type awsInstallStackReader struct {
	client cloudFormationClient
}

type installStackStatus struct {
	Name         string                          `json:"name"`
	Status       string                          `json:"status"`
	Phase        string                          `json:"phase"`
	StatusReason string                          `json:"status_reason,omitempty"`
	StartedAt    *time.Time                      `json:"started_at,omitempty"`
	UpdatedAt    *time.Time                      `json:"updated_at,omitempty"`
	Events       []installStackEvent             `json:"events"`
	Resources    map[string]installStackResource `json:"resources,omitempty"`
}

type installStackResource struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
}

type installStackEvent struct {
	ID                string    `json:"id"`
	LogicalResourceID string    `json:"logical_resource_id,omitempty"`
	ResourceType      string    `json:"resource_type,omitempty"`
	Status            string    `json:"status"`
	StatusReason      string    `json:"status_reason,omitempty"`
	Timestamp         time.Time `json:"timestamp"`
}

func (r *awsInstallStackReader) Read(ctx context.Context, name string) (*installStackStatus, error) {
	described, err := r.client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{StackName: &name})
	if err != nil {
		if stackDoesNotExist(err) {
			return &installStackStatus{Name: name, Status: "NOT_CREATED", Phase: "pending", Events: []installStackEvent{}}, nil
		}
		return nil, fmt.Errorf("describe stack: %w", err)
	}
	if len(described.Stacks) != 1 {
		return nil, fmt.Errorf("describe stack returned %d stacks", len(described.Stacks))
	}
	stack := described.Stacks[0]
	status := string(stack.StackStatus)
	result := &installStackStatus{
		Name: name, Status: status, Phase: installStackPhase(status),
		StatusReason: stringValue(stack.StackStatusReason), StartedAt: stack.CreationTime,
		UpdatedAt: stack.LastUpdatedTime, Events: []installStackEvent{},
	}
	templateOutput, err := r.client.GetTemplate(ctx, &cloudformation.GetTemplateInput{
		StackName: &name, TemplateStage: cloudformationtypes.TemplateStageOriginal,
	})
	if err != nil {
		return nil, fmt.Errorf("read stack template: %w", err)
	}
	var template stackTemplate
	if err := json.Unmarshal([]byte(aws.ToString(templateOutput.TemplateBody)), &template); err != nil {
		return nil, fmt.Errorf("decode stack template: %w", err)
	}
	result.Resources = make(map[string]installStackResource, len(template.Resources))
	for logicalID, resource := range template.Resources {
		result.Resources[logicalID] = installStackResource{Type: resource.Type, Properties: resource.Properties}
	}
	var nextToken *string
	for len(result.Events) < maxInstallStackEvents {
		page, err := r.client.DescribeStackEvents(ctx, &cloudformation.DescribeStackEventsInput{StackName: &name, NextToken: nextToken})
		if err != nil {
			return nil, fmt.Errorf("describe stack events: %w", err)
		}
		for _, event := range page.StackEvents {
			if event.EventId == nil || event.Timestamp == nil {
				continue
			}
			result.Events = append(result.Events, installStackEvent{
				ID: *event.EventId, LogicalResourceID: stringValue(event.LogicalResourceId), ResourceType: stringValue(event.ResourceType),
				Status: string(event.ResourceStatus), StatusReason: stringValue(event.ResourceStatusReason), Timestamp: *event.Timestamp,
			})
			if len(result.Events) == maxInstallStackEvents {
				break
			}
		}
		if page.NextToken == nil || len(result.Events) == maxInstallStackEvents {
			break
		}
		nextToken = page.NextToken
	}
	sort.Slice(result.Events, func(i, j int) bool { return result.Events[i].Timestamp.Before(result.Events[j].Timestamp) })
	if result.StatusReason == "" {
		for i := len(result.Events) - 1; i >= 0; i-- {
			if result.Events[i].StatusReason != "" && installStackPhase(result.Events[i].Status) == "failed" {
				result.StatusReason = result.Events[i].StatusReason
				break
			}
		}
	}
	if result.UpdatedAt == nil && len(result.Events) > 0 {
		result.UpdatedAt = &result.Events[len(result.Events)-1].Timestamp
	}
	return result, nil
}

func installStackPhase(status string) string {
	upper := strings.ToUpper(status)
	switch {
	case upper == "NOT_CREATED":
		return "pending"
	case strings.Contains(upper, "FAILED") || strings.Contains(upper, "ROLLBACK"):
		return "failed"
	case strings.HasSuffix(upper, "_IN_PROGRESS"):
		return "in-progress"
	case strings.HasSuffix(upper, "_COMPLETE"):
		return "finished"
	default:
		return "pending"
	}
}

func stackDoesNotExist(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && apiErr.ErrorCode() == "ValidationError" && strings.Contains(strings.ToLower(apiErr.ErrorMessage()), "does not exist")
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (p *portalServer) installStack(w http.ResponseWriter, r *http.Request) {
	if p.installStackName == "" || p.installStackReader == nil {
		writeRawJSON(w, []byte("null"))
		return
	}
	status, err := p.installStackReader.Read(r.Context(), p.installStackName)
	if err != nil {
		writeAPIError(w, err, http.StatusBadGateway)
		return
	}
	writeJSON(w, status)
}
