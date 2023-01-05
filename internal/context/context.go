package context

import (
	"context"
	"fmt"

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

func GetUserID(ctx context.Context) (string, error) {
	val := ctx.Value(UserIDContext{})
	if val == nil {
		return "", fmt.Errorf("user id not in context")
	}

	userID, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("invalid user id in context: %v", val)
	}

	return userID, nil
}

func GetUser(ctx context.Context) (*models.User, error) {
	val := ctx.Value(UserContext{})
	if val == nil {
		return nil, fmt.Errorf("user not in context")
	}

	user, ok := val.(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid user in context: %v", val)
	}

	return user, nil
}
