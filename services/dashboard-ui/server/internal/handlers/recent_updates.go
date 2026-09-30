package handlers

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const recentUpdatesLimit = 8

var branchRunWorkflowTypes = []string{
	"app_branches_manual_update",
	"app_branches_config_repo_update",
	"app_branches_component_repo_update",
}

type RecentUpdatesHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewRecentUpdatesHandler(cfg *internal.Config, l *zap.Logger) *RecentUpdatesHandler {
	return &RecentUpdatesHandler{cfg: cfg, l: l}
}

func (h *RecentUpdatesHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/recent-updates/sse", h.StreamRecentUpdates)
	return nil
}

func (h *RecentUpdatesHandler) StreamRecentUpdates(c *gin.Context) {
	client, _, ok := sseAuth(c, h.cfg, h.l)
	if !ok {
		return
	}

	state := &recentUpdatesState{
		groupRuns: map[string][]*models.AppInstallGroupRun{},
		details:   map[string]*models.AppWorkflow{},
	}

	runSSEStream(c, sseStreamConfig{
		ClientErrMsg: "failed to fetch recent updates",
		PollInterval: sseOrgStatusPollInterval,
		Log:          h.l,
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			payload, err := fetchRecentUpdates(ctx, client, state, h.l)
			if err != nil {
				return sseFetchResult{}, err
			}
			ev, err := marshalEvent("recent-updates", payload)
			if err != nil {
				return sseFetchResult{}, errors.Join(errSSESilentRetry, err)
			}
			return sseFetchResult{Events: []sseEvent{ev}}, nil
		},
	})
}

type recentRunRef struct {
	workflow   *models.AppWorkflow
	run        *models.AppAppBranchRun
	appID      string
	branchID   string
	branchName string
}

func selectRecentRuns(workflows []*models.AppWorkflow) []recentRunRef {
	byID := map[string]*models.AppWorkflow{}
	for _, workflow := range workflows {
		if workflow == nil || workflow.ID == "" {
			continue
		}
		byID[workflow.ID] = workflow
	}

	refs := make([]recentRunRef, 0, len(byID))
	for _, workflow := range byID {
		ref, ok := recentRunRefFrom(workflow)
		if !ok {
			continue
		}
		refs = append(refs, ref)
	}

	sort.Slice(refs, func(i, j int) bool {
		return recentRunCreatedAt(refs[i]) > recentRunCreatedAt(refs[j])
	})
	if len(refs) > recentUpdatesLimit {
		refs = refs[:recentUpdatesLimit]
	}
	return refs
}

func recentRunRefFrom(workflow *models.AppWorkflow) (recentRunRef, bool) {
	if workflow == nil || len(workflow.AppBranchRuns) == 0 || workflow.AppBranchRuns[0] == nil {
		return recentRunRef{}, false
	}
	run := workflow.AppBranchRuns[0]
	branchID := workflow.OwnerID
	branchName := ""
	appID := ""
	if run.AppBranch != nil {
		if run.AppBranch.ID != "" {
			branchID = run.AppBranch.ID
		}
		branchName = run.AppBranch.Name
		appID = run.AppBranch.AppID
	}
	if branchName == "" {
		branchName = branchID
	}
	if workflow.ID == "" || run.ID == "" || appID == "" || branchID == "" {
		return recentRunRef{}, false
	}
	return recentRunRef{
		workflow:   workflow,
		run:        run,
		appID:      appID,
		branchID:   branchID,
		branchName: branchName,
	}, true
}

func recentRunCreatedAt(ref recentRunRef) string {
	if ref.run != nil && ref.run.CreatedAt != "" {
		return ref.run.CreatedAt
	}
	if ref.workflow != nil {
		return ref.workflow.CreatedAt
	}
	return ""
}

func recentRunStatus(ref recentRunRef) string {
	if ref.run != nil && ref.run.Status != "" {
		return ref.run.Status
	}
	if ref.workflow != nil && ref.workflow.Status != nil {
		return string(ref.workflow.Status.Status)
	}
	return ""
}

type recentUpdatesState struct {
	mu        sync.Mutex
	groupRuns map[string][]*models.AppInstallGroupRun
	details   map[string]*models.AppWorkflow
}

func (s *recentUpdatesState) cachedGroupRuns(runID, status string) ([]*models.AppInstallGroupRun, bool) {
	if !terminalStatuses[status] {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	runs, ok := s.groupRuns[runID]
	return runs, ok
}

func (s *recentUpdatesState) storeGroupRuns(runID, status string, runs []*models.AppInstallGroupRun) {
	if !terminalStatuses[status] {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groupRuns[runID] = runs
}

func (s *recentUpdatesState) cachedDetail(workflowID, status string) (*models.AppWorkflow, bool) {
	if !terminalStatuses[status] {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	workflow, ok := s.details[workflowID]
	return workflow, ok
}

func (s *recentUpdatesState) storeDetail(workflowID, status string, workflow *models.AppWorkflow) {
	if !terminalStatuses[status] || workflow == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.details[workflowID] = workflow
}

func fetchRecentUpdates(ctx context.Context, client nuon.Client, state *recentUpdatesState, l *zap.Logger) (recentUpdatesPayload, error) {
	workflows, err := listBranchRunWorkflows(ctx, client)
	if err != nil {
		return recentUpdatesPayload{}, err
	}
	refs := selectRecentRuns(workflows)
	if len(refs) == 0 {
		return recentUpdatesPayload{Updates: []recentUpdateSource{}}, nil
	}

	approvals, err := client.GetOrgPendingApprovals(ctx)
	if err != nil {
		l.Warn("recent updates approvals", zap.Error(err))
		approvals = nil
	}

	details := map[string]*models.AppWorkflow{}
	configs := map[string]*models.AppAppBranchConfig{}
	apps := map[string]*models.AppApp{}
	installs := map[string][]*models.AppInstall{}
	groupRuns := map[string][]*models.AppInstallGroupRun{}
	var mu sync.Mutex

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	seenBranches := map[string]recentRunRef{}
	seenApps := map[string]recentRunRef{}
	for _, ref := range refs {
		seenBranches[ref.branchID] = ref
		seenApps[ref.appID] = ref
	}

	for _, ref := range seenBranches {
		g.Go(func() error {
			config, err := client.GetAppBranchLatestConfig(gctx, ref.appID, ref.branchID)
			if err != nil {
				l.Warn("recent updates branch config", zap.Error(err))
				return nil
			}
			mu.Lock()
			configs[ref.branchID] = config
			mu.Unlock()
			return nil
		})
		g.Go(func() error {
			planonly := false
			runs, _, err := client.GetAppBranchRunsWithQuery(gctx, ref.appID, ref.branchID, &nuon.GetAppBranchRunsQuery{
				Planonly: &planonly,
				Limit:    recentUpdatesLimit,
			})
			if err != nil {
				l.Warn("recent updates branch runs", zap.Error(err))
				return nil
			}
			mu.Lock()
			for _, workflow := range runs {
				if workflow != nil && workflow.ID != "" {
					details[workflow.ID] = workflow
				}
			}
			mu.Unlock()
			return nil
		})
	}

	for _, ref := range seenApps {
		g.Go(func() error {
			app, err := client.GetApp(gctx, ref.appID)
			if err != nil {
				l.Warn("recent updates app", zap.Error(err))
				return nil
			}
			mu.Lock()
			apps[ref.appID] = app
			mu.Unlock()
			return nil
		})
		g.Go(func() error {
			list, _, err := client.GetAppInstalls(gctx, ref.appID, &models.GetPaginatedQuery{Limit: 100})
			if err != nil {
				l.Warn("recent updates installs", zap.Error(err))
				return nil
			}
			mu.Lock()
			installs[ref.appID] = list
			mu.Unlock()
			return nil
		})
	}

	for _, ref := range refs {
		g.Go(func() error {
			status := recentRunStatus(ref)
			if cached, ok := state.cachedGroupRuns(ref.run.ID, status); ok {
				mu.Lock()
				groupRuns[ref.run.ID] = cached
				mu.Unlock()
				return nil
			}
			runs, err := client.GetInstallGroupRuns(gctx, ref.appID, ref.branchID, ref.run.ID)
			if err != nil {
				l.Warn("recent updates install group runs", zap.Error(err))
				return nil
			}
			state.storeGroupRuns(ref.run.ID, status, runs)
			mu.Lock()
			groupRuns[ref.run.ID] = runs
			mu.Unlock()
			return nil
		})
	}

	_ = g.Wait()

	updates := make([]recentUpdateSource, 0, len(refs))
	for _, ref := range refs {
		detail := details[ref.workflow.ID]
		if detail == nil {
			if cached, ok := state.cachedDetail(ref.workflow.ID, recentRunStatus(ref)); ok {
				detail = cached
			}
		} else {
			state.storeDetail(ref.workflow.ID, recentRunStatus(ref), detail)
		}
		updates = append(updates, buildRecentUpdate(ref, detail, configs[ref.branchID], apps[ref.appID], installs[ref.appID], groupRuns[ref.run.ID], approvals))
	}
	return recentUpdatesPayload{Updates: updates}, nil
}

func listBranchRunWorkflows(ctx context.Context, client nuon.Client) ([]*models.AppWorkflow, error) {
	var mu sync.Mutex
	var workflows []*models.AppWorkflow
	var failures int

	g, gctx := errgroup.WithContext(ctx)
	for _, workflowType := range branchRunWorkflowTypes {
		g.Go(func() error {
			list, err := client.GetOrgWorkflows(gctx, &nuon.GetOrgWorkflowsQuery{
				Planonly: false,
				Type:     workflowType,
				Limit:    recentUpdatesLimit,
			})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failures++
				return nil
			}
			workflows = append(workflows, list...)
			return nil
		})
	}
	_ = g.Wait()
	if failures == len(branchRunWorkflowTypes) {
		return nil, errors.New("unable to list branch run workflows")
	}
	return workflows, nil
}
