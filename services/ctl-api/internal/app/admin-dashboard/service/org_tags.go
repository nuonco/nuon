package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type updateOrgTagsRequest struct {
	Tags []string `json:"tags"`
}

// UpdateOrgTags handles updating tags for an organization
func (s *service) UpdateOrgTags(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	s.l.Info("UpdateOrgTags called", zap.String("org_id", orgID))

	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Org ID is required"})
		return
	}

	var req updateOrgTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.l.Error("failed to parse request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	s.l.Info("Parsed tags", zap.Strings("new_tags", req.Tags))

	// Get current org to compare tags
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Update tags directly in database
	org.Tags = req.Tags

	// Update only the tags field (using Select to only update tags)
	if err := s.db.WithContext(ctx).Model(org).Select("tags").Updates(org).Error; err != nil {
		s.l.Error("failed to update org tags", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tags"})
		return
	}

	// Fetch updated org
	updatedOrg, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get updated org", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated organization"})
		return
	}

	c.JSON(http.StatusOK, updatedOrg)
}

type removeSingleTagRequest struct {
	Tag string `json:"tag"`
}

// RemoveSingleTag handles removing a single tag from an organization
func (s *service) RemoveSingleTag(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	tagToRemove := c.Param("tag")

	s.l.Info("RemoveSingleTag called", zap.String("org_id", orgID), zap.String("tag", tagToRemove))

	if orgID == "" || tagToRemove == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Org ID and tag are required"})
		return
	}

	// Get current org
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Remove the tag from the list
	var newTags []string
	for _, tag := range org.Tags {
		if tag != tagToRemove {
			newTags = append(newTags, tag)
		}
	}

	// Update org with new tags
	org.Tags = newTags

	// Update only the tags field (using Select to only update tags)
	if err := s.db.WithContext(ctx).Model(org).Select("tags").Updates(org).Error; err != nil {
		s.l.Error("failed to update org tags", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove tag"})
		return
	}

	// Fetch updated org
	updatedOrg, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get updated org", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated organization"})
		return
	}

	c.JSON(http.StatusOK, updatedOrg)
}
