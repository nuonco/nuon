package installcreate

import (
	"context"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type latestInputAPI struct {
	nuon.Client
	calls int
}

func (a *latestInputAPI) GetAppInputLatestConfig(context.Context, string) (*models.AppAppInputConfig, error) {
	a.calls++
	return &models.AppAppInputConfig{ID: "input"}, nil
}

func TestResolveInputConfigWithoutBranchDoesNotPoll(t *testing.T) {
	api := &latestInputAPI{}
	got, err := ResolveInputConfig(context.Background(), api, "app", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "input" || api.calls != 1 {
		t.Fatalf("got %v after %d calls", got, api.calls)
	}
}
