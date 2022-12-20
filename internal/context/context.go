package context

import (
	"context"

	"github.com/powertoolsdev/api/internal/models"
)

// UserContext is the context key representing a user.
type UserContext struct{}

// UserIDContext is the context key representing a user id.
type UserIDContext struct{}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDContext{}, userID)
}

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, UserContext{}, user)
}
