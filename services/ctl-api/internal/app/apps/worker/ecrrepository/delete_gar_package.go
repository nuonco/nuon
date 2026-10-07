package ecrrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/artifactregistry/v1"
)

type DeleteGARPackageRequest struct {
	OrgID string `validate:"required" json:"org_id"`
	AppID string `validate:"required" json:"app_id"`
}

func (r DeleteGARPackageRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type DeleteGARPackageResponse struct{}

type garRepository struct {
	location string
	project  string
	name     string
}

func parseGARRepositoryURL(raw string) (*garRepository, error) {
	parts := strings.Split(strings.TrimSuffix(raw, "/"), "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid gar repository url %q: expected <location>-docker.pkg.dev/<project>/<repository>", raw)
	}

	location, ok := strings.CutSuffix(parts[0], "-docker.pkg.dev")
	if !ok || location == "" {
		return nil, fmt.Errorf("invalid gar repository host %q: expected <location>-docker.pkg.dev", parts[0])
	}

	return &garRepository{location: location, project: parts[1], name: parts[2]}, nil
}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 30m
// @start-to-close-timeout 30m
func (a *Activities) DeleteGARPackage(ctx context.Context, req *DeleteGARPackageRequest) (*DeleteGARPackageResponse, error) {
	if err := req.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate request: %w", err)
	}

	repo, err := parseGARRepositoryURL(a.cfg.ManagementGARRepositoryURL)
	if err != nil {
		return nil, err
	}

	svc, err := artifactregistry.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create artifact registry client: %w", err)
	}

	name := fmt.Sprintf(
		"projects/%s/locations/%s/repositories/%s/packages/%s",
		repo.project, repo.location, repo.name, url.PathEscape(req.OrgID+"/"+req.AppID),
	)

	op, found, err := deleteGARPackage(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to delete gar package: %w", err)
	}
	if !found {
		return &DeleteGARPackageResponse{}, nil
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for !op.Done {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}

		op, err = svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to poll gar package deletion: %w", err)
		}
	}

	if op.Error != nil {
		return nil, fmt.Errorf("gar package deletion failed: code %d: %s", op.Error.Code, op.Error.Message)
	}

	return &DeleteGARPackageResponse{}, nil
}

func deleteGARPackage(ctx context.Context, name string) (*artifactregistry.Operation, bool, error) {
	client, err := google.DefaultClient(ctx, artifactregistry.CloudPlatformScope)
	if err != nil {
		return nil, false, fmt.Errorf("unable to create gcp http client: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "https://artifactregistry.googleapis.com/v1/"+name, nil)
	if err != nil {
		return nil, false, fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var op artifactregistry.Operation
	if err := json.NewDecoder(resp.Body).Decode(&op); err != nil {
		return nil, false, fmt.Errorf("unable to parse operation: %w", err)
	}

	return &op, true, nil
}
