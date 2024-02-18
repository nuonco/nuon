package migrations

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration025EnsureCreatedByIDs(ctx context.Context) error {
	var deletedOrgs []*app.Org
	res := a.db.Unscoped().WithContext(ctx).
		Find(&deletedOrgs)
	if res.Error != nil {
		return res.Error
	}

	var orgs []*app.Org
	res = a.db.WithContext(ctx).
		Find(&orgs)
	if res.Error != nil {
		return res.Error
	}

	allOrgs := append(orgs, deletedOrgs...)
	orgCreatedByIDs := make(map[string]string)
	for _, org := range allOrgs {
		orgCreatedByIDs[org.ID] = org.CreatedByID
	}

	// UserTokens
	var userTokens []*app.UserToken
	res = a.db.Unscoped().WithContext(ctx).
		Find(&userTokens)
	if res.Error != nil {
		return res.Error
	}
	for _, userToken := range userTokens {
		if len(userToken.CreatedByID) > 0 {
			continue
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.UserToken{
				ID: userToken.ID,
			}).
			Updates(app.UserToken{
				CreatedByID: userToken.Subject,
			})

		if res.Error != nil {
			return fmt.Errorf("unable to update user tokens: %w", res.Error)
		}
	}

	// Default to me:
	// Sandboxes
	var sandboxes []*app.Sandbox
	res = a.db.Unscoped().WithContext(ctx).
		Find(&sandboxes)
	if res.Error != nil {
		return res.Error
	}
	for _, userToken := range sandboxes {
		if len(userToken.CreatedByID) > 0 {
			continue
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.Sandbox{
				ID: userToken.ID,
			}).
			Updates(app.Sandbox{
				CreatedByID: "google-oauth2|114670241124324496631",
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update sandboxes: %w", res.Error)
		}
	}
	// SandboxReleases
	var sandboxReleases []*app.SandboxRelease
	res = a.db.Unscoped().WithContext(ctx).
		Find(&sandboxReleases)
	if res.Error != nil {
		return res.Error
	}
	for _, userToken := range sandboxReleases {
		if len(userToken.CreatedByID) > 0 {
			continue
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.SandboxRelease{
				ID: userToken.ID,
			}).
			Updates(app.SandboxRelease{
				CreatedByID: "google-oauth2|114670241124324496631",
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update sandboxReleases: %w", res.Error)
		}
	}

	// UserOrg
	var userOrgs []*app.UserOrg
	res = a.db.Unscoped().WithContext(ctx).
		Find(&userOrgs)
	if res.Error != nil {
		return res.Error
	}
	for _, userToken := range userOrgs {
		if len(userToken.CreatedByID) > 0 {
			continue
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.UserOrg{
				ID: userToken.ID,
			}).
			Updates(app.UserOrg{
				CreatedByID: orgCreatedByIDs[userToken.OrgID],
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update userOrgs: %w", res.Error)
		}
	}
	// VCSConnections
	var vcsConns []*app.VCSConnection
	res = a.db.Unscoped().WithContext(ctx).
		Find(&vcsConns)
	if res.Error != nil {
		return res.Error
	}
	for _, userToken := range vcsConns {
		if len(userToken.CreatedByID) > 0 {
			continue
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.VCSConnection{
				ID: userToken.ID,
			}).
			Updates(app.VCSConnection{
				CreatedByID: orgCreatedByIDs[userToken.OrgID],
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update vcsConns: %w", res.Error)
		}
	}
	// VCSConnectionCommits
	var vcsConnCommits []*app.VCSConnectionCommit
	res = a.db.WithContext(ctx).
		Preload("VCSConnection").
		Find(&vcsConnCommits)
	if res.Error != nil {
		return res.Error
	}
	for _, userToken := range vcsConnCommits {
		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.VCSConnectionCommit{
				ID: userToken.ID,
			}).
			Updates(app.VCSConnectionCommit{
				CreatedByID: orgCreatedByIDs[userToken.VCSConnection.OrgID],
				OrgID:       userToken.VCSConnection.OrgID,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update vcsConnCommits: %w", res.Error)
		}
		fmt.Println("updated", userToken.ID, orgCreatedByIDs[userToken.OrgID])
	}

	// ComponentBuilds
	var compBuilds []*app.ComponentBuild
	res = a.db.Unscoped().WithContext(ctx).
		Find(&compBuilds)
	if res.Error != nil {
		return res.Error
	}
	for _, userToken := range compBuilds {
		if len(userToken.CreatedByID) > 0 {
			continue
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.ComponentBuild{
				ID: userToken.ID,
			}).
			Updates(app.ComponentBuild{
				CreatedByID: orgCreatedByIDs[userToken.OrgID],
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update compBuilds: %w", res.Error)
		}
	}
	// InstallDeploys
	var installDeploys []*app.InstallDeploy
	res = a.db.Unscoped().WithContext(ctx).
		Preload("InstallComponent").
		Find(&installDeploys)
	if res.Error != nil {
		return res.Error
	}
	for _, userToken := range installDeploys {
		if len(userToken.CreatedByID) > 0 && len(userToken.OrgID) > 0 {
			continue
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.InstallDeploy{
				ID: userToken.ID,
			}).
			Updates(app.InstallDeploy{
				CreatedByID: orgCreatedByIDs[userToken.InstallComponent.OrgID],
				OrgID:       userToken.InstallComponent.OrgID,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update installDeploys: %w", res.Error)
		}
	}
	// InstallComponents
	var installComponents []*app.InstallComponent
	res = a.db.Unscoped().WithContext(ctx).
		Find(&installComponents)
	if res.Error != nil {
		return res.Error
	}
	for _, userToken := range installComponents {
		if len(userToken.CreatedByID) > 0 {
			continue
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.InstallComponent{
				ID: userToken.ID,
			}).
			Updates(app.InstallComponent{
				CreatedByID: orgCreatedByIDs[userToken.OrgID],
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update installComponents: %w", res.Error)
		}
	}

	return nil
}
