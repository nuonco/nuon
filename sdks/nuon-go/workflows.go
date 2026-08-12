package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type GetInstallWorkflowsQuery struct {
	Finished *bool
	Planonly *bool
	Type     string
	Search   string
	Limit    int
	Offset   int
}

func (c *client) GetInstallWorkflows(ctx context.Context, installID string, query *GetInstallWorkflowsQuery) ([]*models.AppWorkflow, bool, error) {
	params := &operations.GetWorkflowsParams{
		InstallID: installID,
		Context:   ctx,
	}

	var limit, offset int
	if query != nil {
		params.Finished = query.Finished
		params.Planonly = query.Planonly
		if query.Type != "" {
			params.Type = &query.Type
		}
		if query.Search != "" {
			params.Search = &query.Search
		}
		limit = query.Limit
		offset = query.Offset
	}
	if limit == 0 {
		limit = 10
	}
	l := int64(limit)
	o := int64(offset)
	params.Limit = &l
	params.Offset = &o

	hr := newResponseHeaderReader(&operations.GetWorkflowsReader{})
	resp, err := c.genClient.Operations.GetWorkflows(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetWorkflows(ctx context.Context, installID string, query *models.GetPaginatedQuery) ([]*models.AppWorkflow, bool, error) {
	params := &operations.GetWorkflowsParams{
		InstallID: installID,
		Context:   ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetWorkflowsReader{})
	resp, err := c.genClient.Operations.GetWorkflows(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetWorkflow(ctx context.Context, owner WorkflowOwner, workflowID string) (*models.AppWorkflow, error) {
	switch {
	case owner.ownedByInstall():
		resp, err := c.genClient.Operations.GetWorkflowByInstall(&operations.GetWorkflowByInstallParams{
			InstallID:  owner.InstallID,
			WorkflowID: workflowID,
			Context:    ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	case owner.ownedByAppBranch():
		resp, err := c.genClient.Operations.GetWorkflowByAppBranch(&operations.GetWorkflowByAppBranchParams{
			AppID:       owner.AppID,
			AppBranchID: owner.AppBranchID,
			WorkflowID:  workflowID,
			Context:     ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	}

	resp, err := c.genClient.Operations.GetWorkflow(&operations.GetWorkflowParams{
		WorkflowID: workflowID,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// CancelWorkflow returns the bare route's Accepted type for every owner so
// callers keep one result shape; the nested responses carry no body either.
func (c *client) CancelWorkflow(ctx context.Context, owner WorkflowOwner, workflowID string) (*operations.CancelWorkflowAccepted, error) {
	switch {
	case owner.ownedByInstall():
		if _, err := c.genClient.Operations.CancelWorkflowByInstall(&operations.CancelWorkflowByInstallParams{
			InstallID:  owner.InstallID,
			WorkflowID: workflowID,
			Context:    ctx,
		}, c.getOrgIDAuthInfo()); err != nil {
			return nil, err
		}
		return operations.NewCancelWorkflowAccepted(), nil
	case owner.ownedByAppBranch():
		if _, err := c.genClient.Operations.CancelWorkflowByAppBranch(&operations.CancelWorkflowByAppBranchParams{
			AppID:       owner.AppID,
			AppBranchID: owner.AppBranchID,
			WorkflowID:  workflowID,
			Context:     ctx,
		}, c.getOrgIDAuthInfo()); err != nil {
			return nil, err
		}
		return operations.NewCancelWorkflowAccepted(), nil
	}

	resp, err := c.genClient.Operations.CancelWorkflow(&operations.CancelWorkflowParams{
		WorkflowID: workflowID,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *client) UpdateWorkflow(ctx context.Context, owner WorkflowOwner, workflowID string, req *models.ServiceUpdateWorkflowRequest) (*models.AppWorkflow, error) {
	switch {
	case owner.ownedByInstall():
		resp, err := c.genClient.Operations.UpdateWorkflowByInstall(&operations.UpdateWorkflowByInstallParams{
			InstallID:  owner.InstallID,
			WorkflowID: workflowID,
			Req:        req,
			Context:    ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	case owner.ownedByAppBranch():
		resp, err := c.genClient.Operations.UpdateWorkflowByAppBranch(&operations.UpdateWorkflowByAppBranchParams{
			AppID:       owner.AppID,
			AppBranchID: owner.AppBranchID,
			WorkflowID:  workflowID,
			Req:         req,
			Context:     ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	}

	resp, err := c.genClient.Operations.UpdateWorkflow(&operations.UpdateWorkflowParams{
		WorkflowID: workflowID,
		Req:        req,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// GetWorkflowSteps returns every step for a workflow. The endpoint is
// paginated (max 100 per page), so this pages through the full set following
// the X-Nuon-Page-Next header rather than loading everything in one query.
func (c *client) GetWorkflowSteps(ctx context.Context, owner WorkflowOwner, workflowID string) ([]*models.AppWorkflowStep, error) {
	const pageLimit = int64(100)

	var (
		steps  []*models.AppWorkflowStep
		offset int64
	)
	for {
		limit := pageLimit
		off := offset

		var (
			page []*models.AppWorkflowStep
			more bool
			err  error
		)
		switch {
		case owner.ownedByInstall():
			hr := newResponseHeaderReader(&operations.GetWorkflowStepsByInstallReader{})
			var resp *operations.GetWorkflowStepsByInstallOK
			resp, err = c.genClient.Operations.GetWorkflowStepsByInstall(&operations.GetWorkflowStepsByInstallParams{
				InstallID:  owner.InstallID,
				WorkflowID: workflowID,
				Limit:      &limit,
				Offset:     &off,
				Context:    ctx,
			}, c.getOrgIDAuthInfo(), hr.ClientOption())
			if resp != nil {
				page, more = resp.Payload, hasNextPage(hr)
			}
		case owner.ownedByAppBranch():
			hr := newResponseHeaderReader(&operations.GetWorkflowStepsByAppBranchReader{})
			var resp *operations.GetWorkflowStepsByAppBranchOK
			resp, err = c.genClient.Operations.GetWorkflowStepsByAppBranch(&operations.GetWorkflowStepsByAppBranchParams{
				AppID:       owner.AppID,
				AppBranchID: owner.AppBranchID,
				WorkflowID:  workflowID,
				Limit:       &limit,
				Offset:      &off,
				Context:     ctx,
			}, c.getOrgIDAuthInfo(), hr.ClientOption())
			if resp != nil {
				page, more = resp.Payload, hasNextPage(hr)
			}
		default:
			hr := newResponseHeaderReader(&operations.GetWorkflowStepsReader{})
			var resp *operations.GetWorkflowStepsOK
			resp, err = c.genClient.Operations.GetWorkflowSteps(&operations.GetWorkflowStepsParams{
				WorkflowID: workflowID,
				Limit:      &limit,
				Offset:     &off,
				Context:    ctx,
			}, c.getOrgIDAuthInfo(), hr.ClientOption())
			if resp != nil {
				page, more = resp.Payload, hasNextPage(hr)
			}
		}
		if err != nil {
			return nil, err
		}

		steps = append(steps, page...)

		if len(page) == 0 || !more {
			break
		}
		offset += pageLimit
	}

	return steps, nil
}

func (c *client) GetWorkflowStep(ctx context.Context, owner WorkflowOwner, workflowID, stepID string) (*models.AppWorkflowStep, error) {
	switch {
	case owner.ownedByInstall():
		resp, err := c.genClient.Operations.GetWorkflowStepByInstall(&operations.GetWorkflowStepByInstallParams{
			InstallID:  owner.InstallID,
			WorkflowID: workflowID,
			StepID:     stepID,
			Context:    ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	case owner.ownedByAppBranch():
		resp, err := c.genClient.Operations.GetWorkflowStepByAppBranch(&operations.GetWorkflowStepByAppBranchParams{
			AppID:       owner.AppID,
			AppBranchID: owner.AppBranchID,
			WorkflowID:  workflowID,
			StepID:      stepID,
			Context:     ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	}

	resp, err := c.genClient.Operations.GetWorkflowStep(&operations.GetWorkflowStepParams{
		WorkflowID: workflowID,
		StepID:     stepID,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) RetryWorkflowStep(ctx context.Context, owner WorkflowOwner, workflowID, stepID string, req *models.ServiceRetryWorkflowStepRequest) error {
	// Note: req parameter is ignored in the current API - the endpoint no longer accepts a request body
	switch {
	case owner.ownedByInstall():
		_, err := c.genClient.Operations.RetryWorkflowStepByInstall(&operations.RetryWorkflowStepByInstallParams{
			InstallID:  owner.InstallID,
			WorkflowID: workflowID,
			StepID:     stepID,
			Context:    ctx,
		}, c.getOrgIDAuthInfo())
		return err
	case owner.ownedByAppBranch():
		_, err := c.genClient.Operations.RetryWorkflowStepByAppBranch(&operations.RetryWorkflowStepByAppBranchParams{
			AppID:       owner.AppID,
			AppBranchID: owner.AppBranchID,
			WorkflowID:  workflowID,
			StepID:      stepID,
			Context:     ctx,
		}, c.getOrgIDAuthInfo())
		return err
	}

	_, err := c.genClient.Operations.RetryWorkflowStep(&operations.RetryWorkflowStepParams{
		WorkflowID: workflowID,
		StepID:     stepID,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	return err
}
