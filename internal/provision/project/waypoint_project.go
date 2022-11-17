package project

import (
	"context"

	"github.com/go-playground/validator/v10"
)

type CreateWaypointProjectRequest struct{}

func (r CreateWaypointProjectRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateWaypointProjectResponse struct{}

func (a *Activities) CreateWaypointProject(ctx context.Context, req CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
	var resp CreateWaypointProjectResponse
	return resp, nil
}
