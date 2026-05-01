package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type addOrgLabelsRequest struct {
	Labels map[string]string `json:"labels"`
}

func (s *service) AddOrgLabels(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")

	var req addOrgLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "org not found"})
		return
	}

	org.Labels.Merge(labels.Labels(req.Labels))

	if err := s.db.WithContext(ctx).Model(org).Select("labels").Updates(org).Error; err != nil {
		s.l.Error("failed to update org labels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update labels: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, org)
}

type removeOrgLabelRequest struct {
	Key string `json:"key"`
}

func (s *service) RemoveOrgLabel(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	key := c.Param("key")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "org not found"})
		return
	}

	delete(org.Labels, key)

	if err := s.db.WithContext(ctx).
		Model(&app.Org{}).
		Where("id = ?", orgID).
		Update("labels", org.Labels).Error; err != nil {
		s.l.Error("failed to remove org label", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to remove label: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, org)
}
