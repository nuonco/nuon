package installs

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestRequireInstallRegion(t *testing.T) {
	for _, cloud := range []string{"AWS", "Azure", "GCP"} {
		t.Run(cloud, func(t *testing.T) {
			require.EqualError(t, requireInstallRegion("", cloud), "--region is required for "+cloud+" installs")
			require.NoError(t, requireInstallRegion("us-west-2", cloud))
		})
	}
}

func TestParseInstallInputs(t *testing.T) {
	inputs, err := parseInstallInputs([]string{"token=part=two", "empty="})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"token": "part=two", "empty": ""}, inputs)

	for _, input := range []string{"missing-separator", "=missing-name"} {
		t.Run(input, func(t *testing.T) {
			_, err := parseInstallInputs([]string{input})
			require.EqualError(t, err, `invalid input "`+input+`": expected name=value`)
		})
	}
}

type branchListAPI struct {
	nuon.Client
	branches []*models.AppAppBranch
}

func (a *branchListAPI) GetAppBranches(context.Context, string, *models.GetPaginatedQuery) ([]*models.AppAppBranch, bool, error) {
	return a.branches, false, nil
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}

func TestResolveCreateInstallBranchPrintsNoticeWithoutAppBranchID(t *testing.T) {
	api := &branchListAPI{branches: []*models.AppAppBranch{{ID: "b1", Name: "prod"}}}
	s := &Service{api: api, cfg: &config.Config{Interactive: false}}

	var branch *createInstallBranch
	stderr := captureStderr(t, func() {
		var err error
		branch, err = s.resolveCreateInstallBranch(context.Background(), "app-1", "", "", true)
		require.NoError(t, err)
	})

	assert.Equal(t, "", branch.branchID)
	assert.Contains(t, stderr, "no --app-branch-id; using the app config from the last nuon apps sync")
}

func TestResolveCreateInstallBranchOmitsNoticeWithoutBranches(t *testing.T) {
	api := &branchListAPI{}
	s := &Service{api: api, cfg: &config.Config{Interactive: false}}

	stderr := captureStderr(t, func() {
		_, err := s.resolveCreateInstallBranch(context.Background(), "app-1", "", "", true)
		require.NoError(t, err)
	})

	assert.Equal(t, "", stderr)
}

func TestResolveCreateInstallBranchOmitsNoticeWithAppBranchID(t *testing.T) {
	branch := &models.AppAppBranch{ID: "b1", Name: "prod"}
	api := &branchListAPI{branches: []*models.AppAppBranch{branch}}
	s := &Service{api: api, cfg: &config.Config{Interactive: false}}

	stderr := captureStderr(t, func() {
		got, err := s.resolveCreateInstallBranch(context.Background(), "app-1", "b1", "", true)
		require.NoError(t, err)
		assert.Equal(t, "b1", got.branchID)
	})

	assert.Equal(t, "", stderr)
}
