package syncer

import (
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	testseedconfig "github.com/nuonco/nuon/services/ctl-api/tests/testseed/config"
)

// A helper that bypasses the sync transaction fails the foreign key instead of seeing the new rows.
func (s *SyncFieldsTestSuite) TestNewBranchGetsAConfig() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Branches = []*config.AppBranchConfig{
		{
			Name: "main",
			PublicRepo: &config.PublicRepoConfig{
				Repo:      "https://github.com/nuonco/example",
				Directory: "/",
				Branch:    "main",
			},
		},
	}

	ctx, testApp, _ := s.sync(cfg)

	var branch app.AppBranch
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.AppBranch{AppID: testApp.ID, Name: "main"}).
		First(&branch).Error, "branch must be created")

	var branchCfg app.AppBranchConfig
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Preload("PublicGitVCSConfig").
		Where(app.AppBranchConfig{AppBranchID: branch.ID}).
		First(&branchCfg).Error, "branch config must be created")

	s.Require().NotNil(branchCfg.PublicGitVCSConfig)
	s.Equal("https://github.com/nuonco/example", branchCfg.PublicGitVCSConfig.Repo)
}

func (s *SyncFieldsTestSuite) TestNewComponentDependenciesArePersisted() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	base := testseedconfig.BuildJobComponent("base")
	dependent := testseedconfig.BuildJobComponent("dependent")
	dependent.Dependencies = []string{"base"}
	cfg.Components = config.ComponentList{base, dependent}

	ctx, testApp, _ := s.sync(cfg)

	var baseComp, dependentComp app.Component
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.Component{AppID: testApp.ID, Name: "base"}).First(&baseComp).Error)
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.Component{AppID: testApp.ID, Name: "dependent"}).First(&dependentComp).Error)

	var deps []app.ComponentDependency
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.ComponentDependency{ComponentID: dependentComp.ID}).
		Find(&deps).Error)

	s.Require().Len(deps, 1, "dependency on a component created in the same sync must be persisted")
	s.Equal(baseComp.ID, deps[0].DependencyID)
}
