package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// UpdateOrgTags handles the form submission from the selectbox component
func (s *service) UpdateOrgTags(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	s.l.Info("UpdateOrgTags called", zap.String("org_id", orgID))

	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Org ID is required"})
		return
	}

	// Parse form data - checkboxes send multiple values with the same name
	if err := c.Request.ParseForm(); err != nil {
		s.l.Error("failed to parse form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Get all tags from the form (checkboxes send multiple values with name="tags")
	newTags := c.Request.Form["tags"]
	s.l.Info("Parsed form tags", zap.Strings("new_tags", newTags), zap.Any("form", c.Request.Form))

	// Get current org to compare tags
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Update tags directly in database
	org.Tags = newTags

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

	// Render the updated org header with popover
	component := views.OrgHeaderWithPopover(updatedOrg)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
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

	// Render the updated org header with popover
	component := views.OrgHeaderWithPopover(updatedOrg)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
