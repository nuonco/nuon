package lookup

import (
	"context"
	"errors"
	"testing"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type installAPI struct {
	nuon.Client
	install *models.AppInstall
	err     error
	gotID   string
}

func (a *installAPI) GetInstall(_ context.Context, installID string) (*models.AppInstall, error) {
	a.gotID = installID
	if a.err != nil {
		return nil, a.err
	}
	return a.install, nil
}

type statusErr struct {
	code int
}

func (e statusErr) Error() string                         { return "api error" }
func (e statusErr) IsCode(code int) bool                  { return e.code == code }
func (e statusErr) IsServerError() bool                   { return e.code >= 500 }
func (e statusErr) GetPayload() *models.StderrErrResponse { return &models.StderrErrResponse{} }

func TestInstallIDNotFound(t *testing.T) {
	t.Parallel()

	_, err := InstallID(context.Background(), &installAPI{err: statusErr{code: 404}}, "inst_missing")
	var userErr *ui.CLIUserError
	if !errors.As(err, &userErr) {
		t.Fatalf("expected CLIUserError, got %v", err)
	}
	if userErr.Msg != `install "inst_missing" not found` {
		t.Fatalf("unexpected message: %q", userErr.Msg)
	}
}

func TestInstallIDResolvesExisting(t *testing.T) {
	t.Parallel()

	got, err := InstallID(context.Background(), &installAPI{
		install: &models.AppInstall{ID: "inst_resolved"},
	}, "my-install")
	if err != nil {
		t.Fatal(err)
	}
	if got != "inst_resolved" {
		t.Fatalf("got %q", got)
	}
}
