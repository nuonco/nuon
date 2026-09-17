package lookup

import (
	"context"
	"errors"
	"testing"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type branchAPI struct {
	nuon.Client
	branches []*models.AppAppBranch
	err      error
}

func (a *branchAPI) GetAppBranches(_ context.Context, _ string) ([]*models.AppAppBranch, error) {
	if a.err != nil {
		return nil, a.err
	}
	return a.branches, nil
}

func TestAppBranchIDResolvesNameOrID(t *testing.T) {
	t.Parallel()

	api := &branchAPI{
		branches: []*models.AppAppBranch{
			{ID: "apb_123", Name: "preview"},
		},
	}

	got, err := AppBranchID(context.Background(), api, "app_1", "preview")
	if err != nil {
		t.Fatal(err)
	}
	if got != "apb_123" {
		t.Fatalf("name lookup: got %q", got)
	}

	got, err = AppBranchID(context.Background(), api, "app_1", "apb_123")
	if err != nil {
		t.Fatal(err)
	}
	if got != "apb_123" {
		t.Fatalf("id lookup: got %q", got)
	}
}

func TestAppBranchIDNotFound(t *testing.T) {
	t.Parallel()

	_, err := AppBranchID(context.Background(), &branchAPI{}, "app_1", "missing")
	var userErr *ui.CLIUserError
	if !errors.As(err, &userErr) {
		t.Fatalf("expected CLIUserError, got %v", err)
	}
	if userErr.Msg != `branch "missing" not found` {
		t.Fatalf("unexpected message: %q", userErr.Msg)
	}
}
