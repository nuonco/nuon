package dev

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go/models"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
	"github.com/powertoolsdev/mono/pkg/config/validate"
)

const (
	defaultSyncTimeout           time.Duration = time.Minute * 12
	defaultSyncSleep             time.Duration = time.Second * 20
	componentBuildStatusError                  = "error"
	componentBuildStatusBuilding               = "building"
	componentBuildStatusActive                 = "active"
	componentStatusQueued                      = "queued"
)

// Dev syncs, buids, and deploys app changes to a dev install.
// It does a few pre-flight checks to make sure we're ready to deploy, then executes the sync, builds, and deploys.
func (s *Service) Dev(ctx context.Context, dir, installID string, autoApprove bool) error {
	var err error
	defer func() {
		fmt.Println()
		if err != nil {
			bubbles.PrintStyledError("Dev workflow failed")
		} else {
			bubbles.PrintStyledSuccess("Dev workflow complete")
		}
	}()

	s.autoApprove = autoApprove

	//
	// Pre-flight checks
	//

	fmt.Println()
	bubbles.PrintStyledInfo("Checking that you are ready to create a new app version...")
	fmt.Println()

	ui.PrintLn("checking git branch...")
	branchName, err := s.checkGitBranch(ctx)
	if err != nil {
		return ui.PrintError(err)
	}
	ui.PrintSuccess(fmt.Sprintf("you are on branch %s", branchName))

	ui.PrintLn("verifying app exists...")
	appID, err := s.getApp(ctx, dir)
	if err != nil {
		return ui.PrintError(err)
	}
	ui.PrintSuccess(fmt.Sprintf("app ID is %s", appID))

	ui.PrintLn("parsing app config...")
	cfg, err := parse.ParseDir(ctx, parse.ParseConfig{
		Dirname: dir,
		V:       validator.New(),
		FileProcessor: func(name string, contents map[string]any) map[string]any {
			if strings.HasPrefix(name, "components/") {
				cmpName := strings.Split(strings.Split(name, "/")[1], ".")[0]
				repoType := ""
				switch {
				case contents["connected_repo"] != nil:
					repoType = "connected_repo"
				case contents["public_repo"] != nil:
					repoType = "public_repo"
				}
				if repoType == "" {
					return contents
				}

				cmpBranch := contents[repoType].(map[string]any)["branch"]
				if cmpBranch != branchName {
					if err := prompt(s.autoApprove, "component %s is on branch %s. Override with %s?", cmpName, cmpBranch, branchName); err == nil {
						contents[repoType].(map[string]any)["branch"] = branchName
					}
				}
			}
			return contents
		},
	})
	if err != nil {
		return ui.PrintError(err)
	}
	ui.PrintSuccess("parsed app config successfully")

	ui.PrintLn("validating config...")
	err = validate.Validate(ctx, s.v, cfg)
	if err != nil {
		if config.IsWarningErr(err) {
			ui.PrintError(err)
		} else {
			return ui.PrintError(err)
		}
	}
	ui.PrintSuccess("app config is valid")

	ui.PrintLn("checking component branches...")
	_, err = s.checkComponentBranches(ctx, cfg, branchName)
	if err != nil {
		return ui.PrintError(err)
	}
	ui.PrintSuccess("component branches are ok")

	ui.PrintLn("checking that local changes have been pushed...")
	err = s.checkLocalChanges(ctx)
	if err != nil {
		return ui.PrintError(err)
	}
	ui.PrintSuccess("all required changes have been pushed")

	//
	// Create new app version
	//

	if err := prompt(autoApprove, "You are ready to create a new version of your app. Continue?"); err != nil {
		return ui.PrintError(err)
	}

	ui.PrintLn("syncing config to api...")
	syncer := sync.New(s.api, appID, "", cfg)
	err = syncer.Sync(ctx)
	if err != nil {
		return ui.PrintError(err)
	}
	ui.PrintSuccess(fmt.Sprintf("app config version %s created", syncer.GetAppConfigID()))
	s.notifyOrphanedComponents(syncer.OrphanedComponents())
	s.notifyOrphanedActions(syncer.OrphanedActions())
	cmpsScheduled := syncer.GetComponentsScheduled()

	if len(cmpsScheduled) > 0 {
		ui.PrintLn("some components require new builds...")
		fmt.Println()
		err = s.pollComponentBuilds(ctx, cmpsScheduled)
		if err != nil {
			return ui.PrintError(err)
		}
		ui.PrintSuccess("builds complete")
	}

	//
	// Deploy new app version
	//
	if installID == "" {
		return ui.PrintError(errors.New("No install is selected. Please select an install to deploy to."))
	}
	if err := prompt(autoApprove, "Ready to deploy the new app config version. Deploy to %s?", installID); err != nil {
		return ui.PrintError(err)
	}

	ui.PrintLn("updating install to use new app config version...")
	err = s.api.UpdateAppConfigInstalls(ctx, appID, syncer.GetAppConfigID(), &models.ServiceUpdateAppConfigInstallsRequest{
		UpdateAll:  false,
		InstallIDs: []string{installID},
	})
	if err != nil {
		return ui.PrintError(err)
	}
	ui.PrintSuccess(fmt.Sprintf("install is updated to use app config version %s", syncer.GetAppConfigID()))

	ui.PrintLn("deploying changes...")
	deploys := []*models.AppInstallDeploy{}
	for _, comp := range cmpsScheduled {
		build, err := s.api.GetComponentLatestBuild(ctx, comp.ID)
		if err != nil {
			return ui.PrintJSONError(err)
		}

		deploy, err := s.api.CreateInstallDeploy(ctx, installID, &models.ServiceCreateInstallDeployRequest{
			BuildID: build.ID,
		})
		if err != nil {
			return ui.PrintJSONError(err)
		}
		deploy.ComponentName = comp.Name
		deploys = append(deploys, deploy)
	}
	fmt.Println()
	err = s.pollDeploys(ctx, installID, deploys)
	if err != nil {
		return ui.PrintError(err)
	}
	ui.PrintSuccess("all changes have been deployed")

	return nil
}
