package services

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/request/middleware"
)

var (
	errInvalidContext = fmt.Errorf("user was not set in context")
)

func getCurrentUser(ctx context.Context) (*models.User, error) {
	obj := ctx.Value(middleware.UserContext{})
	if obj == nil {
		return nil, errInvalidContext
	}

	user, ok := obj.(*models.User)
	if !ok {
		return nil, errInvalidContext
	}

	return user, nil
}
