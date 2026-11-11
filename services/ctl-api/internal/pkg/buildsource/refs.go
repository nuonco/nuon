package buildsource

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
)

func ResolveRef(ctx context.Context, db *gorm.DB, vcs *vcshelpers.Helpers, connected *app.ConnectedGithubVCSConfig, public *app.PublicGitVCSConfig, runID, ref string) (string, error) {
	if sourceidentity.IsCommit(ref) {
		return ref, nil
	}
	repository, orgID, scope := "", "", "public"
	if connected != nil {
		repository, orgID, scope = connected.Repo, connected.OrgID, connected.VCSConnectionID
		if ref == "" {
			ref = connected.Branch
		}
	} else if public != nil {
		repository, orgID = public.Repo, public.OrgID
		if ref == "" {
			ref = public.Branch
		}
	} else {
		return "", fmt.Errorf("source ref has no repository config")
	}
	resolve := func() (string, error) {
		if connected != nil {
			commit, err := vcs.GetConnectedGithubVCSConfigCommit(ctx, connected, ref)
			if err != nil {
				return "", err
			}
			if commit != nil && sourceidentity.IsCommit(commit.GetSHA()) {
				return commit.GetSHA(), nil
			}
		} else {
			commit, err := vcs.GetPublicGitVCSConfigCommit(ctx, public, ref)
			if err != nil {
				return "", err
			}
			if commit != nil && sourceidentity.IsCommit(commit.GetSHA()) {
				return commit.GetSHA(), nil
			}
		}
		return "", fmt.Errorf("source did not resolve to a full commit SHA")
	}
	if runID == "" {
		return resolve()
	}
	key := "source-ref:" + sourceidentity.Hash([]string{orgID, scope, normalizeRepo(repository), ref})
	pinned := func() (string, error) {
		var run app.AppBranchRun
		if err := db.WithContext(ctx).Select("id", "org_id", "labels").Where(app.AppBranchRun{ID: runID}).First(&run).Error; err != nil {
			return "", err
		}
		if run.OrgID != orgID {
			return "", fmt.Errorf("source config does not belong to branch run org")
		}
		pin := run.Labels[key]
		if pin != "" && !sourceidentity.IsCommit(pin) {
			return "", fmt.Errorf("invalid pinned source ref")
		}
		return pin, nil
	}
	if pin, err := pinned(); err != nil || pin != "" {
		return pin, err
	}
	sha, err := resolve()
	if err != nil {
		return "", err
	}
	if err := db.WithContext(ctx).Model(&app.AppBranchRun{}).
		Where(app.AppBranchRun{ID: runID}).
		Where("COALESCE(labels->>?, '') = ''", key).
		UpdateColumn("labels", gorm.Expr("COALESCE(labels, '{}'::jsonb) || jsonb_build_object(?::text, ?::text)", key, sha)).Error; err != nil {
		return "", err
	}
	pin, err := pinned()
	if err == nil && pin == "" {
		err = fmt.Errorf("source ref pin was not recorded")
	}
	return pin, err
}
