package installs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func testBuild(id, createdAt, branchID string, status models.AppStatus) *models.AppComponentBuild {
	return &models.AppComponentBuild{
		ID:          id,
		CreatedAt:   createdAt,
		AppBranchID: branchID,
		StatusV2:    &models.AppCompositeStatus{Status: status},
	}
}

func TestLatestDeployableBuildID(t *testing.T) {
	t.Run("picks newest active build over the one already deployed", func(t *testing.T) {
		builds := []*models.AppComponentBuild{
			testBuild("bldold", "2026-10-20T10:00:00Z", "", models.AppStatusActive),
			testBuild("bldnew", "2026-11-23T10:00:00Z", "", models.AppStatusActive),
		}
		id, ok := latestDeployableBuildID(builds, "")
		require.True(t, ok)
		assert.Equal(t, "bldnew", id)
	})

	t.Run("compares fractional seconds as times, not strings", func(t *testing.T) {
		builds := []*models.AppComponentBuild{
			testBuild("bldshort", "2026-11-23T10:00:00.12Z", "", models.AppStatusActive),
			testBuild("bldlong", "2026-11-23T10:00:00.123456Z", "", models.AppStatusActive),
		}
		id, ok := latestDeployableBuildID(builds, "")
		require.True(t, ok)
		assert.Equal(t, "bldlong", id)
	})

	t.Run("skips builds with an unparseable timestamp", func(t *testing.T) {
		builds := []*models.AppComponentBuild{
			testBuild("bldbad", "not-a-time", "", models.AppStatusActive),
			testBuild("bldok", "2026-11-23T10:00:00Z", "", models.AppStatusActive),
		}
		id, ok := latestDeployableBuildID(builds, "")
		require.True(t, ok)
		assert.Equal(t, "bldok", id)
	})

	t.Run("skips failed, in-progress, and preview builds", func(t *testing.T) {
		preview := testBuild("bldpreview", "2026-11-23T12:00:00Z", "", models.AppStatusActive)
		preview.IsPreview = true
		builds := []*models.AppComponentBuild{
			testBuild("bldfailed", "2026-11-23T11:00:00Z", "", models.AppStatusError),
			testBuild("bldbuilding", "2026-11-23T11:30:00Z", "", models.AppStatusBuilding),
			preview,
			{ID: "bldnostatus", CreatedAt: "2026-11-23T12:30:00Z"},
			testBuild("bldgood", "2026-11-22T10:00:00Z", "", models.AppStatusActive),
		}
		id, ok := latestDeployableBuildID(builds, "")
		require.True(t, ok)
		assert.Equal(t, "bldgood", id)
	})

	t.Run("only uses builds on the install's app branch", func(t *testing.T) {
		otherBranchRun := testBuild("bldrun", "2026-11-23T11:00:00Z", "", models.AppStatusActive)
		otherBranchRun.AppBranchRun = &models.AppAppBranchRun{AppBranch: &models.AppAppBranch{ID: "abrother"}}
		sameBranchRun := testBuild("bldmainrun", "2026-11-21T10:00:00Z", "", models.AppStatusActive)
		sameBranchRun.AppBranchRun = &models.AppAppBranchRun{AppBranch: &models.AppAppBranch{ID: "abrmain"}}
		builds := []*models.AppComponentBuild{
			testBuild("bldother", "2026-11-23T12:00:00Z", "abrother", models.AppStatusActive),
			otherBranchRun,
			{ID: "bldnobranch", CreatedAt: "2026-11-23T13:00:00Z", StatusV2: &models.AppCompositeStatus{Status: models.AppStatusActive}},
			sameBranchRun,
			testBuild("bldmain", "2026-11-20T10:00:00Z", "abrmain", models.AppStatusActive),
		}
		id, ok := latestDeployableBuildID(builds, "abrmain")
		require.True(t, ok)
		assert.Equal(t, "bldmainrun", id)
	})

	t.Run("no deployable build", func(t *testing.T) {
		id, ok := latestDeployableBuildID([]*models.AppComponentBuild{
			nil,
			testBuild("bldfailed", "2026-11-23T11:00:00Z", "", models.AppStatusError),
		}, "")
		require.False(t, ok)
		assert.Empty(t, id)
	})
}
