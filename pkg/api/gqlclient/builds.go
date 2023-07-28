package gqlclient

import (
	"context"
	"fmt"
)

type Build struct {
	buildFields
}

func (c *client) CancelBuild(ctx context.Context, buildID string) (bool, error) {
	cancelResp, err := cancelBuild(ctx, c.graphqlClient, buildID)
	if err != nil {
		return false, fmt.Errorf("unable to cancel build: %w", err)
	}
	return cancelResp.CancelBuild, nil
}

func (c *client) StartBuild(ctx context.Context, input BuildInput) (*Build, error) {
	startResp, err := startBuild(ctx, c.graphqlClient, &input)
	if err != nil {
		return nil, fmt.Errorf("unable to start build: %w", err)
	}
	return &Build{
		startResp.StartBuild.buildFields,
	}, nil
}

func (c *client) GetBuild(ctx context.Context, buildID string) (*Build, error) {
	buildResp, err := getBuild(ctx, c.graphqlClient, buildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get build: %w", err)
	}
	return &Build{buildResp.Build.buildFields}, nil
}

func (c *client) GetBuilds(ctx context.Context, componentID string) ([]*Build, error) {
	buildsResp, err := getBuilds(ctx, c.graphqlClient, componentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get builds: %w", err)
	}

	builds := make([]*Build, 0)
	for _, build := range buildsResp.Builds {
		builds = append(builds, &Build{
			build.buildFields,
		})
	}

	return builds, nil
}

func (c *client) GetBuildStatus(ctx context.Context, buildID string) (Status, error) {
	buildResp, err := getBuild(ctx, c.graphqlClient, buildID)
	if err != nil {
		return StatusUnspecified, fmt.Errorf("unable to get build: %w", err)
	}

	compResp, err := getComponent(ctx, c.graphqlClient, buildResp.Build.ComponentId)
	if err != nil {
		return StatusUnspecified, fmt.Errorf("unable to get component: %w", err)
	}

	statusResp, err := getBuildStatus(ctx, c.graphqlClient,
		compResp.Component.App.Id,
		buildID,
		compResp.Component.Id,
		compResp.Component.App.Org.Id)

	if err != nil {
		return StatusUnspecified, fmt.Errorf("unable to upsert install: %w", err)
	}

	return statusResp.BuildStatus, nil
}
