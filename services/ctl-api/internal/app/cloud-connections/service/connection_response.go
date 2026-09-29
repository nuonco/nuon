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
	Installs int64 `json:"installs"`
}

func userError(err error) error {
	return stderr.ErrUser{Err: err, Description: err.Error()}
}

func (s *service) response(ctx context.Context, connection *app.CloudConnection) (ConnectionResponse, error) {
	usage, err := s.usage(ctx, connection.ID)
	if err != nil {
		return ConnectionResponse{}, err
	}
	response := ConnectionResponse{CloudConnection: *connection, Setup: s.setup(connection), UsedBy: usage, VerificationInProgress: verificationInProgress(connection)}
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

func (s *service) usage(ctx context.Context, connectionID string) (ConnectionUsage, error) {
	var usage ConnectionUsage
	if err := s.db.WithContext(ctx).Model(&app.Install{}).Where(app.Install{CloudConnectionID: &connectionID}).Count(&usage.Installs).Error; err != nil {
		return usage, fmt.Errorf("count installs using cloud connection: %w", err)
	}
	return usage, nil
}

func (s *service) usages(ctx context.Context, connectionIDs []string) (map[string]ConnectionUsage, error) {
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
		Where("cloud_connection_id IN ?", connectionIDs).
		Group("cloud_connection_id").
		Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("count installs using cloud connections: %w", err)
	}
	for _, count := range counts {
		usage[count.CloudConnectionID] = ConnectionUsage{Installs: count.Installs}
	}
	return usage, nil
}
