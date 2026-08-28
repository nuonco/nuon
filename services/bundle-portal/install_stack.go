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

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operationstate"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
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
	ID           string                          `json:"id,omitempty"`
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
		ID: aws.ToString(stack.StackId), Name: name, Status: status, Phase: installStackPhase(status),
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

func (p *portalServer) installationRegistration(w http.ResponseWriter, r *http.Request) {
	registration, err := p.buildInstallationRegistration(r.Context())
	if err != nil {
		writeAPIError(w, err, registrationErrorStatus(err))
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="nuon-installation-registration-%s.json"`, registration.DeploymentID))
	writeJSON(w, registration)
}

func (p *portalServer) buildInstallationRegistration(ctx context.Context) (customermanaged.InstallationRegistration, error) {
	if p.deploymentID == "" || p.cloudProvider == "" || p.cloudAccountID == "" || p.cloudRegion == "" || p.installStackReader == nil {
		return customermanaged.InstallationRegistration{}, fmt.Errorf("installation registration is unavailable for this portal deployment")
	}
	registrationRaw, found, err := p.store.Get(ctx, customermanaged.InstallationRegistrationKey)
	if err != nil {
		return customermanaged.InstallationRegistration{}, err
	}
	if found {
		var registration customermanaged.InstallationRegistration
		if err := json.Unmarshal(registrationRaw, &registration); err != nil {
			return customermanaged.InstallationRegistration{}, fmt.Errorf("decode installation registration: %w", err)
		}
		if err := registration.Validate(); err != nil {
			return customermanaged.InstallationRegistration{}, fmt.Errorf("validate installation registration: %w", err)
		}
		return registration, nil
	}
	bundle, err := p.initialBundle(ctx)
	if err != nil {
		return customermanaged.InstallationRegistration{}, err
	}
	if bundle.Release == nil || bundle.Package == nil {
		return customermanaged.InstallationRegistration{}, fmt.Errorf("active bundle does not identify its release and package")
	}
	statusRaw, found, err := p.store.Get(ctx, "status.json")
	if err != nil {
		return customermanaged.InstallationRegistration{}, err
	}
	if !found {
		return customermanaged.InstallationRegistration{}, fmt.Errorf("installation has not finished")
	}
	var status statestore.Status
	if err := json.Unmarshal(statusRaw, &status); err != nil {
		return customermanaged.InstallationRegistration{}, fmt.Errorf("decode installation status: %w", err)
	}
	if status.Status != statestore.RunStatusFinished || status.FinishedAt == nil || bundle.ActivatedAt.IsZero() {
		return customermanaged.InstallationRegistration{}, fmt.Errorf("installation has not finished")
	}
	installRun, err := p.initialInstallRun(ctx, status, bundle.BundleDigest)
	if err != nil {
		return customermanaged.InstallationRegistration{}, err
	}
	stack, err := p.installStackReader.Read(ctx, p.installStackName)
	if err != nil {
		return customermanaged.InstallationRegistration{}, fmt.Errorf("read installation stack: %w", err)
	}
	if stack.ID == "" || stack.Phase != "finished" {
		return customermanaged.InstallationRegistration{}, fmt.Errorf("installation stack has not finished")
	}
	registration, err := customermanaged.NewInstallationRegistration(customermanaged.InstallationRegistration{
		SchemaVersion: customermanaged.InstallationRegistrationSchemaVersion,
		ReleaseID:     bundle.Release.ID, ReleaseDigest: bundle.Release.Digest,
		PackageID: bundle.Package.ID, PackageDigest: bundle.Package.Digest,
		BundleDigest: bundle.BundleDigest, ArchiveDigest: bundle.ArchiveDigest,
		OperationID:  installRun.RunID,
		DeploymentID: p.deploymentID, InstallID: status.InstallID,
		Cloud:       customermanaged.InstallationRegistrationCloud{Provider: p.cloudProvider, AccountID: p.cloudAccountID, Region: p.cloudRegion},
		Stack:       customermanaged.InstallationRegistrationStack{Type: "aws-cloudformation", ID: stack.ID, Name: stack.Name},
		InstalledAt: *installRun.FinishedAt,
	})
	if err != nil {
		return customermanaged.InstallationRegistration{}, err
	}
	registrationRaw, err = json.Marshal(registration)
	if err != nil {
		return customermanaged.InstallationRegistration{}, fmt.Errorf("encode installation registration: %w", err)
	}
	if err := p.controlStore.PutIfAbsent(ctx, customermanaged.InstallationRegistrationKey, registrationRaw); err != nil && !errors.Is(err, operationstate.ErrObjectExists) {
		return customermanaged.InstallationRegistration{}, fmt.Errorf("persist installation registration: %w", err)
	}
	return registration, nil

}

func (p *portalServer) initialInstallRun(ctx context.Context, current statestore.Status, bundleDigest string) (statestore.Status, error) {
	keys, err := p.store.List(ctx, statestore.InstallRunsPrefix)
	if err != nil {
		return statestore.Status{}, err
	}
	var initial statestore.Status
	for _, key := range keys {
		if !strings.HasSuffix(key, "/status.json") {
			continue
		}
		raw, found, err := p.store.Get(ctx, key)
		if err != nil {
			return statestore.Status{}, err
		}
		if !found {
			continue
		}
		var status statestore.Status
		if err := json.Unmarshal(raw, &status); err != nil {
			return statestore.Status{}, fmt.Errorf("decode install run status: %w", err)
		}
		if status.RunType != statestore.RunTypeInstall || status.Status != statestore.RunStatusFinished ||
			status.FinishedAt == nil || !strings.EqualFold(status.BundleDigest, bundleDigest) {
			continue
		}
		if initial.FinishedAt == nil || status.FinishedAt.Before(*initial.FinishedAt) {
			initial = status
		}
	}
	if initial.FinishedAt != nil {
		return initial, nil
	}
	if current.RunType == "" || current.RunType == statestore.RunTypeInstall {
		return current, nil
	}
	return statestore.Status{}, fmt.Errorf("initial installation completion not found")
}

func (p *portalServer) initialBundle(ctx context.Context) (operation.BundleInfo, error) {
	keys, err := p.store.List(ctx, operation.BundlesPrefix)
	if err != nil {
		return operation.BundleInfo{}, err
	}
	var initial operation.BundleInfo
	for _, key := range keys {
		if !strings.HasSuffix(key, ".json") {
			continue
		}
		raw, found, err := p.store.Get(ctx, key)
		if err != nil {
			return operation.BundleInfo{}, err
		}
		if !found {
			continue
		}
		var bundle operation.BundleInfo
		if err := json.Unmarshal(raw, &bundle); err != nil {
			return operation.BundleInfo{}, fmt.Errorf("decode bundle history: %w", err)
		}
		if bundle.Release == nil || bundle.Package == nil || bundle.ActivatedAt.IsZero() ||
			(!initial.ActivatedAt.IsZero() && !bundle.ActivatedAt.Before(initial.ActivatedAt)) {
			continue
		}
		initial = bundle
	}
	if !initial.ActivatedAt.IsZero() {
		return initial, nil
	}
	raw, found, err := p.store.Get(ctx, operation.BundleKey)
	if err != nil {
		return operation.BundleInfo{}, err
	}
	if !found {
		return operation.BundleInfo{}, fmt.Errorf("active bundle not found")
	}
	if err := json.Unmarshal(raw, &initial); err != nil {
		return operation.BundleInfo{}, fmt.Errorf("decode active bundle: %w", err)
	}
	return initial, nil
}

func registrationErrorStatus(err error) int {
	if strings.Contains(err.Error(), "read installation stack") {
		return http.StatusBadGateway
	}
	if strings.Contains(err.Error(), "decode") {
		return http.StatusInternalServerError
	}
	return http.StatusConflict
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
