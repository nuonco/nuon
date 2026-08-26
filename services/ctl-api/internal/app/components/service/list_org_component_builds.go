package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	componentBuildCursorOlder = "older"
	componentBuildCursorNewer = "newer"
)

type componentBuildCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Direction string    `json:"direction"`
}

type OrgComponentBuildHistoryItem struct {
	Build            app.ComponentBuild `json:"build"`
	AppID            string             `json:"app_id"`
	ComponentID      string             `json:"component_id"`
	ComponentName    string             `json:"component_name"`
	BuildRunnerJobID *string            `json:"build_runner_job_id"`
}

type OrgComponentBuildHistoryResponse struct {
	Items          []OrgComponentBuildHistoryItem `json:"items"`
	NextCursor     *string                        `json:"next_cursor"`
	PreviousCursor *string                        `json:"previous_cursor"`
}

// @ID ListOrgComponentBuilds
// @Summary list component build history for the current organization
// @Tags components
// @Produce json
// @Security APIKey && OrgID
// @Param limit query int false "limit of builds to return" Default(10)
// @Param cursor query string false "opaque component build history cursor"
// @Success 200 {object} OrgComponentBuildHistoryResponse
// @Failure 400 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/component-builds [get]
func (s *service) ListOrgComponentBuilds(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid limit %q", ctx.Query("limit")),
			Description: "limit must be between 1 and 100",
		})
		return
	}

	cursor, err := decodeComponentBuildCursor(ctx.Query("cursor"))
	if err != nil {
		ctx.Error(stderr.ErrUser{Err: err, Description: "invalid component build history cursor"})
		return
	}

	builds, hasMore, err := s.listOrgComponentBuilds(ctx, org.ID, cursor, limit)
	if err != nil {
		ctx.Error(err)
		return
	}

	buildPtrs := make([]*app.ComponentBuild, len(builds))
	for i := range builds {
		buildPtrs[i] = &builds[i]
	}
	if err := s.hydrateBuildRunnerJobs(ctx, buildPtrs...); err != nil {
		ctx.Error(err)
		return
	}

	response, err := buildOrgComponentBuildHistoryResponse(builds, cursor, hasMore)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *service) listOrgComponentBuilds(ctx *gin.Context, orgID string, cursor *componentBuildCursor, limit int) ([]app.ComponentBuild, bool, error) {
	builds := []app.ComponentBuild{}
	query := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("VCSConnectionCommit").
		Preload("ComponentConfigConnection.Component").
		Where(&app.ComponentBuild{OrgID: orgID})

	newer := cursor != nil && cursor.Direction == componentBuildCursorNewer
	if cursor != nil {
		query = query.Where(componentBuildCursorClause(cursor))
	}
	if newer {
		query = query.
			Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
			Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}})
	} else {
		query = query.
			Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
			Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true})
	}

	if err := query.Limit(limit + 1).Find(&builds).Error; err != nil {
		return nil, false, fmt.Errorf("unable to list organization component builds: %w", err)
	}

	hasMore := len(builds) > limit
	if hasMore {
		builds = builds[:limit]
	}
	if newer {
		reverseComponentBuilds(builds)
	}
	return builds, hasMore, nil
}

func componentBuildCursorClause(cursor *componentBuildCursor) clause.Expression {
	createdAt := clause.Column{Name: "created_at"}
	id := clause.Column{Name: "id"}
	if cursor.Direction == componentBuildCursorNewer {
		return clause.Or(
			clause.Gt{Column: createdAt, Value: cursor.CreatedAt},
			clause.And(
				clause.Eq{Column: createdAt, Value: cursor.CreatedAt},
				clause.Gt{Column: id, Value: cursor.ID},
			),
		)
	}
	return clause.Or(
		clause.Lt{Column: createdAt, Value: cursor.CreatedAt},
		clause.And(
			clause.Eq{Column: createdAt, Value: cursor.CreatedAt},
			clause.Lt{Column: id, Value: cursor.ID},
		),
	)
}

func buildOrgComponentBuildHistoryResponse(builds []app.ComponentBuild, cursor *componentBuildCursor, hasMore bool) (*OrgComponentBuildHistoryResponse, error) {
	response := &OrgComponentBuildHistoryResponse{
		Items: make([]OrgComponentBuildHistoryItem, 0, len(builds)),
	}
	for i := range builds {
		build := builds[i]
		response.Items = append(response.Items, OrgComponentBuildHistoryItem{
			Build:            build,
			AppID:            build.ComponentConfigConnection.Component.AppID,
			ComponentID:      build.ComponentID,
			ComponentName:    build.ComponentName,
			BuildRunnerJobID: build.BuildRunnerJobID,
		})
	}
	if len(builds) == 0 {
		return response, nil
	}

	newer := cursor != nil && cursor.Direction == componentBuildCursorNewer
	if ((cursor == nil || !newer) && hasMore) || newer {
		next, err := encodeComponentBuildCursor(builds[len(builds)-1], componentBuildCursorOlder)
		if err != nil {
			return nil, err
		}
		response.NextCursor = &next
	}
	if cursor != nil && (!newer || hasMore) {
		previous, err := encodeComponentBuildCursor(builds[0], componentBuildCursorNewer)
		if err != nil {
			return nil, err
		}
		response.PreviousCursor = &previous
	}
	return response, nil
}

func encodeComponentBuildCursor(build app.ComponentBuild, direction string) (string, error) {
	data, err := json.Marshal(componentBuildCursor{
		CreatedAt: build.CreatedAt,
		ID:        build.ID,
		Direction: direction,
	})
	if err != nil {
		return "", fmt.Errorf("unable to encode component build cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeComponentBuildCursor(value string) (*componentBuildCursor, error) {
	if value == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}
	var cursor componentBuildCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("parse cursor: %w", err)
	}
	if cursor.CreatedAt.IsZero() || cursor.ID == "" || (cursor.Direction != componentBuildCursorOlder && cursor.Direction != componentBuildCursorNewer) {
		return nil, fmt.Errorf("cursor is incomplete")
	}
	return &cursor, nil
}

func reverseComponentBuilds(builds []app.ComponentBuild) {
	for left, right := 0, len(builds)-1; left < right; left, right = left+1, right-1 {
		builds[left], builds[right] = builds[right], builds[left]
	}
}
