package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ListCommits returns recent commits for a VCS connection.
func (s *service) ListCommits(ctx *gin.Context) {
	connectionID := ctx.Param("connection_id")
	if connectionID == "" {
		ctx.Error(fmt.Errorf("connection_id is required"))
		return
	}

	limit := 25
	if l := ctx.Query("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n <= 0 {
			ctx.Error(fmt.Errorf("invalid limit: %s", l))
			return
		}
		if n > 200 {
			n = 200
		}
		limit = n
	}

	var commits []app.VCSConnectionCommit
	query := s.db.WithContext(ctx).
		Where("owner_id = ? AND owner_type = ?", connectionID, "vcs_connections").
		Order("created_at DESC").
		Limit(limit)

	if branch := ctx.Query("branch"); branch != "" {
		query = query.Where("branch = ?", branch)
	}

	if err := query.Find(&commits).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to list commits: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, commits)
}
