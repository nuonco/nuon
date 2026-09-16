package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const appBranchesPerPage = 20

// appBranchOwnerType is the polymorphic owner type app branches are recorded
// under on queues, workflows and signals.
const appBranchOwnerType = "app_branches"

type appBranchFilters struct {
	search      string
	managedBy   string
	showDeleted bool
}

func (s *service) AppBranches(c *gin.Context) {
	ctx := c.Request.Context()
	filters := appBranchFilters{
		search:      c.Query("search"),
		managedBy:   c.Query("managed_by"),
		showDeleted: c.Query("show_deleted") == "true",
	}
	sort := c.DefaultQuery("sort", "newest")
	page := getPageFromQuery(c)

	branches, totalPages, err := s.getAppBranches(ctx, filters, sort, page)
	if err != nil {
		s.l.Error("failed to get app branches", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch app branches"})
		return
	}

	branchViews := make([]views.AppBranchView, 0, len(branches))
	for _, branch := range branches {
		branchViews = append(branchViews, views.AppBranchView{
			AppBranch: branch,
			OrgName:   branch.Org.Name,
			AppName:   branch.App.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"app_branches": branchViews,
		"page":         page,
		"total_pages":  totalPages,
	})
}

func (s *service) appBranchQuery(ctx context.Context, filters appBranchFilters) *gorm.DB {
	query := s.readDB().WithContext(ctx).Model(&app.AppBranch{})

	if filters.showDeleted {
		query = query.Unscoped()
	}

	if search := strings.TrimSpace(filters.search); search != "" {
		switch {
		case strings.HasPrefix(search, "abr"):
			query = query.Where(&app.AppBranch{ID: search})
		case strings.HasPrefix(search, "app"):
			query = query.Where(&app.AppBranch{AppID: search})
		case strings.HasPrefix(search, "org"):
			query = query.Where(&app.AppBranch{OrgID: search})
		default:
			query = query.Where("app_branches.name ILIKE ?", "%"+search+"%")
		}
	}

	if filters.managedBy != "" {
		query = query.Where(&app.AppBranch{ManagedBy: app.AppBranchManagedBy(filters.managedBy)})
	}

	return query
}

func (s *service) getAppBranches(
	ctx context.Context,
	filters appBranchFilters,
	sortOrder string,
	page int,
) ([]*app.AppBranch, int, error) {
	var totalCount int64
	if err := s.appBranchQuery(ctx, filters).Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count app branches: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(appBranchesPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	orderClause := "app_branches.created_at DESC"
	if sortOrder == "oldest" {
		orderClause = "app_branches.created_at ASC"
	}

	offset := (page - 1) * appBranchesPerPage
	var branches []*app.AppBranch
	if err := s.appBranchQuery(ctx, filters).
		Select(fmt.Sprintf("app_branches.*, "+
			"(SELECT COUNT(*) FROM %s w "+
			"WHERE w.owner_type = '%s' AND w.owner_id = app_branches.id AND w.deleted_at = 0) AS workflow_count",
			(&app.Workflow{}).TableName(), appBranchOwnerType)).
		Preload("Org").
		Preload("App").
		Order(orderClause).
		Offset(offset).
		Limit(appBranchesPerPage).
		Find(&branches).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to get app branches: %w", err)
	}

	if err := s.attachLatestAppBranchRuns(ctx, branches); err != nil {
		return nil, 0, fmt.Errorf("unable to get latest app branch runs: %w", err)
	}

	return branches, totalPages, nil
}

// attachLatestAppBranchRuns sets LatestRun on each branch with a single query.
func (s *service) attachLatestAppBranchRuns(ctx context.Context, branches []*app.AppBranch) error {
	if len(branches) == 0 {
		return nil
	}

	branchIDs := make([]string, 0, len(branches))
	for _, branch := range branches {
		branchIDs = append(branchIDs, branch.ID)
	}

	latestRunIDs := s.readDB().WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Select("DISTINCT ON (app_branch_id) id").
		Where("app_branch_id IN ?", branchIDs).
		Order("app_branch_id, created_at DESC")

	var runs []app.AppBranchRun
	if err := s.readDB().WithContext(ctx).
		Where("id IN (?)", latestRunIDs).
		Find(&runs).Error; err != nil {
		return err
	}

	runsByBranchID := make(map[string]*app.AppBranchRun, len(runs))
	for i := range runs {
		runsByBranchID[runs[i].AppBranchID] = &runs[i]
	}

	for _, branch := range branches {
		branch.LatestRun = runsByBranchID[branch.ID]
	}

	return nil
}
