package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type ConnectionResponse struct {
	app.CloudConnection
	Setup                  SetupResponse   `json:"setup"`
	UsedBy                 ConnectionUsage `json:"used_by"`
	VerificationInProgress bool            `json:"verification_in_progress"`
}

type ConnectionListResponse struct {
	app.CloudConnection
	UsedBy                 ConnectionUsage `json:"used_by"`
	VerificationInProgress bool            `json:"verification_in_progress"`
}

type ConnectionUsage struct {
	Installs          int64 `json:"installs"`
	ImageRepositories int64 `json:"image_repositories"`
}

func userError(err error) error {
	return stderr.ErrUser{Err: err, Description: err.Error()}
}

func (s *service) response(ctx context.Context, connection *app.CloudConnection, repositories ...string) (ConnectionResponse, error) {
	usage, err := s.usages(ctx, connection.OrgID, []string{connection.ID})
	if err != nil {
		return ConnectionResponse{}, err
	}
	response := ConnectionResponse{CloudConnection: *connection, Setup: s.setup(connection, repositories...), UsedBy: usage[connection.ID], VerificationInProgress: verificationInProgress(connection)}
	applyVerificationRequestedAt(&response.CloudConnection)
	return response, nil
}

func newListResponse(connection *app.CloudConnection, usage map[string]ConnectionUsage) ConnectionListResponse {
	response := ConnectionListResponse{CloudConnection: *connection, UsedBy: usage[connection.ID], VerificationInProgress: verificationInProgress(connection)}
	applyVerificationRequestedAt(&response.CloudConnection)
	return response
}

func applyVerificationRequestedAt(connection *app.CloudConnection) {
	if connection.VerificationRequestedAt != nil {
		requestedAt := connection.VerificationRequestedAt.UTC()
		connection.VerificationRequestedAt = &requestedAt
	}
}

func verificationInProgress(connection *app.CloudConnection) bool {
	return connection.VerificationRequestedAt != nil && (connection.LastVerifiedAt == nil || connection.VerificationRequestedAt.After(*connection.LastVerifiedAt))
}

func (s *service) usages(ctx context.Context, orgID string, connectionIDs []string) (map[string]ConnectionUsage, error) {
	usage := make(map[string]ConnectionUsage, len(connectionIDs))
	if len(connectionIDs) == 0 {
		return usage, nil
	}
	var counts []struct {
		CloudConnectionID string
		Installs          int64
	}
	if err := s.db.WithContext(ctx).Model(&app.Install{}).
		Select("cloud_connection_id, count(*) as installs").
		Where(app.Install{OrgID: orgID}).
		Where("cloud_connection_id IN ?", connectionIDs).
		Group("cloud_connection_id").
		Find(&counts).Error; err != nil {
		return nil, fmt.Errorf("count installs using cloud connections: %w", err)
	}
	for _, count := range counts {
		usage[count.CloudConnectionID] = ConnectionUsage{Installs: count.Installs}
	}
	var repositoryCounts []struct {
		CloudConnectionID string
		ImageRepositories int64
	}
	if err := s.db.WithContext(ctx).Model(&app.ImageRepository{}).
		Select("cloud_connection_id, count(*) as image_repositories").
		Where(app.ImageRepository{OrgID: orgID}).
		Where("cloud_connection_id IN ?", connectionIDs).
		Group("cloud_connection_id").
		Find(&repositoryCounts).Error; err != nil {
		return nil, fmt.Errorf("count image repositories using cloud connections: %w", err)
	}
	for _, count := range repositoryCounts {
		value := usage[count.CloudConnectionID]
		value.ImageRepositories = count.ImageRepositories
		usage[count.CloudConnectionID] = value
	}
	return usage, nil
}
