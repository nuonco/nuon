package apps

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// branchNameResolver translates between the names a branch config file uses and
// the IDs the API speaks. Both directions are needed: names go out on create
// requests, and IDs coming back have to be rendered as names again so an
// unchanged config compares equal to what is already deployed.
type branchNameResolver struct {
	api   nuon.Client
	appID string

	installsLoaded  bool
	installNameToID map[string]string
	installIDToName map[string]string

	runbookNameToID map[string]string
	runbookIDToName map[string]string
}

func newBranchNameResolver(api nuon.Client, appID string) *branchNameResolver {
	return &branchNameResolver{
		api:             api,
		appID:           appID,
		installNameToID: map[string]string{},
		installIDToName: map[string]string{},
		runbookNameToID: map[string]string{},
		runbookIDToName: map[string]string{},
	}
}

func (r *branchNameResolver) loadInstalls(ctx context.Context) error {
	if r.installsLoaded {
		return nil
	}

	var (
		offset  int
		limit   = 50
		hasMore = true
	)
	for hasMore {
		installs, more, err := r.api.GetAppInstalls(ctx, r.appID, &models.GetPaginatedQuery{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			return fmt.Errorf("unable to list installs for app %s: %w", r.appID, err)
		}
		for _, install := range installs {
			if install.Name == "" {
				continue
			}
			r.installNameToID[install.Name] = install.ID
			r.installIDToName[install.ID] = install.Name
		}
		offset += limit
		hasMore = more
	}

	r.installsLoaded = true
	return nil
}

func (r *branchNameResolver) installID(ctx context.Context, name string) (string, error) {
	if err := r.loadInstalls(ctx); err != nil {
		return "", err
	}
	id, ok := r.installNameToID[name]
	if !ok {
		return "", fmt.Errorf("unknown install name: %s", name)
	}
	return id, nil
}

// installName returns the install's name, or an empty string when the ID no
// longer resolves (a deleted install), so callers can fall back to the raw ID.
func (r *branchNameResolver) installName(ctx context.Context, id string) (string, error) {
	if err := r.loadInstalls(ctx); err != nil {
		return "", err
	}
	return r.installIDToName[id], nil
}

func (r *branchNameResolver) runbookID(ctx context.Context, name string) (string, error) {
	if id, ok := r.runbookNameToID[name]; ok {
		return id, nil
	}
	runbook, err := r.api.GetAppRunbook(ctx, r.appID, name)
	if err != nil {
		return "", fmt.Errorf("unknown post_deploy_runbooks runbook name %q: %w", name, err)
	}
	r.runbookNameToID[name] = runbook.ID
	r.runbookIDToName[runbook.ID] = runbook.Name
	return runbook.ID, nil
}

// runbookName maps a runbook ID back to its name, falling back to the ID when
// the runbook has since been deleted.
func (r *branchNameResolver) runbookName(ctx context.Context, id string) string {
	if name, ok := r.runbookIDToName[id]; ok {
		return name
	}
	runbook, err := r.api.GetAppRunbook(ctx, r.appID, id)
	if err != nil || runbook.Name == "" {
		r.runbookIDToName[id] = id
		return id
	}
	r.runbookIDToName[id] = runbook.Name
	r.runbookNameToID[runbook.Name] = runbook.ID
	return runbook.Name
}
