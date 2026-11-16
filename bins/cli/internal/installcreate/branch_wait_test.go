package installcreate

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type fakeBranchConfigAPI struct {
	nuon.Client
	polls   []*models.AppAppBranch
	cfg     *models.AppAppConfig
	fetches int
}

func (a *fakeBranchConfigAPI) GetAppBranch(context.Context, string, string) (*models.AppAppBranch, error) {
	a.fetches++
	if a.fetches >= len(a.polls) {
		return a.polls[len(a.polls)-1], nil
	}
	return a.polls[a.fetches], nil
}

func (a *fakeBranchConfigAPI) GetAppConfig(context.Context, string, string, *bool) (*models.AppAppConfig, error) {
	return a.cfg, nil
}

func TestAwaitBranchConfigFailsWhenLatestOfSeveralRunsFailed(t *testing.T) {
	api := &fakeBranchConfigAPI{}
	branch := &models.AppAppBranch{
		ID:   "branch-1",
		Name: "feature",
		LatestRun: &models.AppAppBranchRun{
			Status: "failed",
		},
	}

	_, err := AwaitBranchConfig(context.Background(), api, "app-acme", branch, false, nil)
	if err == nil || err.Error() != `app branch "feature" run failed` {
		t.Fatalf("err = %v", err)
	}
	if api.fetches != 0 {
		t.Fatalf("expected no polling, got %d fetches", api.fetches)
	}
}

func TestAwaitBranchConfigFailsWhenLatestRunWasSkipped(t *testing.T) {
	api := &fakeBranchConfigAPI{}
	branch := &models.AppAppBranch{
		ID:   "branch-1",
		Name: "feature",
		LatestRun: &models.AppAppBranchRun{
			Status: "not-attempted",
		},
	}

	_, err := AwaitBranchConfig(context.Background(), api, "app-acme", branch, false, nil)
	if err == nil || err.Error() != `app branch "feature" latest run was skipped; run the branch again` {
		t.Fatalf("err = %v", err)
	}
	if api.fetches != 0 {
		t.Fatalf("expected no polling, got %d fetches", api.fetches)
	}
}

func TestAwaitBranchConfigWaitsWhileLatestRunIsInFlight(t *testing.T) {
	origPoll, origTimeout := BranchConfigPoll, BranchConfigTimeout
	BranchConfigPoll = time.Millisecond
	BranchConfigTimeout = 20 * time.Millisecond
	t.Cleanup(func() {
		BranchConfigPoll = origPoll
		BranchConfigTimeout = origTimeout
	})

	api := &fakeBranchConfigAPI{
		polls: []*models.AppAppBranch{nil, {
			ID:   "branch-1",
			Name: "feature",
			LatestRun: &models.AppAppBranchRun{
				Status: "in-progress",
			},
		}},
	}
	branch := &models.AppAppBranch{
		ID:   "branch-1",
		Name: "feature",
		LatestRun: &models.AppAppBranchRun{
			Status: "in-progress",
		},
	}

	_, err := AwaitBranchConfig(context.Background(), api, "app-acme", branch, false, nil)
	if err == nil || !strings.Contains(err.Error(), "has no app config to apply") {
		t.Fatalf("err = %v", err)
	}
	if api.fetches == 0 {
		t.Fatalf("expected the in-flight run to be polled at least once")
	}
}
