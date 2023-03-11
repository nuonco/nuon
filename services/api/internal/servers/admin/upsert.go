package admin

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	adminv1 "github.com/powertoolsdev/mono/pkg/types/api/admin/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

func (s *server) UpsertSandboxVersion(ctx context.Context, req *connect.Request[adminv1.UpsertSandboxVersionRequest]) (*connect.Response[adminv1.UpsertSandboxVersionResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	sandboxVersion, err := s.Svc.UpsertSandboxVersion(ctx, models.SandboxVersionInput{
		ID:             req.Msg.Id,
		SandboxName:    req.Msg.SandboxName,
		SandboxVersion: req.Msg.SandboxVersion,
		TfVersion:      req.Msg.TfVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to upsert sandbox version: %w", err)
	}

	return connect.NewResponse(&adminv1.UpsertSandboxVersionResponse{
		Sandbox: &adminv1.SandboxVersion{
			Id:             sandboxVersion.GetID(),
			SandboxName:    sandboxVersion.SandboxName,
			SandboxVersion: sandboxVersion.SandboxVersion,
			TfVersion:      sandboxVersion.TfVersion,
		},
	}), nil
}
