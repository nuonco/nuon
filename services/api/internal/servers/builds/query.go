package builds

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	buildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

func (s *server) GetBuild(
	ctx context.Context,
	req *connect.Request[buildv1.GetBuildRequest],
) (*connect.Response[buildv1.GetBuildResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// retrieve build from DB
	var build models.Build
	if err := s.db.WithContext(ctx).First(&build, "id = ?", req.Msg.Id).Error; err != nil {
		return nil, fmt.Errorf("retrieving build failed: %w", err)
	}

	return connect.NewResponse(&buildv1.GetBuildResponse{
		Build: build.ToProto(),
	}), nil
}

func (s *server) QueryBuilds(
	ctx context.Context,
	req *connect.Request[buildv1.QueryBuildsRequest],
) (*connect.Response[buildv1.QueryBuildsResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// use gorm model to retrieve builds
	rows, err := s.db.Model(&models.Build{}).Where("component_id = ?", req.Msg.ComponentId).Rows()
	if err != nil {
		return nil, fmt.Errorf("retrieving builds failed: %w", err)
	}
	defer rows.Close()

	// iterate through rows and convert to proto
	builds := []*buildv1.Build{}
	for rows.Next() {
		var build *models.Build
		s.db.ScanRows(rows, &build)
		builds = append(builds, build.ToProto())
	}

	return connect.NewResponse(&buildv1.QueryBuildsResponse{
		Builds: builds,
	}), nil
}

func (s *server) ListBuildsByInstance(
	ctx context.Context,
	req *connect.Request[buildv1.ListBuildsByInstanceRequest],
) (*connect.Response[buildv1.ListBuildsByInstanceResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// use gorm model to retrieve builds
	rows, err := s.db.Model(&models.Build{}).Where("instance_id = ?", req.Msg.InstanceId).Rows()
	if err != nil {
		return nil, fmt.Errorf("retrieving builds failed: %w", err)
	}
	defer rows.Close()

	// iterate through rows and convert to proto
	builds := []*buildv1.Build{}
	for rows.Next() {
		var build *models.Build
		s.db.ScanRows(rows, &build)
		builds = append(builds, build.ToProto())
	}

	return connect.NewResponse(&buildv1.ListBuildsByInstanceResponse{
		Builds: builds,
	}), nil
}
