package examples

//go:generate go run ../main.go generate .

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

// SimpleActivity is a basic activity with minimal configuration
// @temporal-gen-v2 activity
func SimpleActivity(ctx context.Context, input string) (string, error) {
	return "result", nil
}

// ComplexActivity demonstrates all available activity options
// @temporal-gen-v2 activity
// @schedule-to-close-timeout 1h
// @start-to-close-timeout 30m
// @max-retries 5
func ComplexActivity(ctx context.Context, input int) (int, error) {
	return input * 2, nil
}

type User struct {
	ID   string
	Name string
}

// ActivityWithCustomOptions demonstrates callback options
// @temporal-gen-v2 activity
// @options-callback GetUserActivityOptions
func ActivityWithCustomOptions(ctx context.Context, user User) error {
	return nil
}

func GetUserActivityOptions(user User) *workflow.ActivityOptions {
	return &workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
}

// ActivityWithByID demonstrates the by-id pattern
// @temporal-gen-v2 activity
// @by-field ID
func ActivityWithByID(ctx context.Context, user User) error {
	return nil
}

// ActivityWithByIDOnly demonstrates the by-id-only pattern
// @temporal-gen-v2 activity
// @by-field ID
// @by-field-only
func ActivityWithByIDOnly(ctx context.Context, user User) error {
	return nil
}

type Activities struct{}

// myPrivateActivity demonstrates the wrapper pattern
// @temporal-gen-v2 activity
// @as-wrapper
func (a *Activities) myPrivateActivity(ctx context.Context, userID string, count int) (string, error) {
	return "done", nil
}

// myWrapperWithByField
// @temporal-gen-v2 activity
// @as-wrapper
// @by-field UserID
func (a *Activities) myWrapperWithByField(ctx context.Context, userID string, count int) (string, error) {
	return "done", nil
}
