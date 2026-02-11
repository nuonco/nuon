package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// UpdateOrgTags handles the form submission from the tags input component
func (s *service) UpdateOrgTags(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	s.l.Info("UpdateOrgTags called", zap.String("org_id", orgID))

	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Org ID is required"})
		return
	}

	// Parse form data - tagsinput sends multiple values with the same name
	if err := c.Request.ParseForm(); err != nil {
		s.l.Error("failed to parse form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Get all tags from the form (tagsinput sends multiple hidden inputs with name="tags")
	newTags := c.Request.Form["tags"]
	s.l.Info("Parsed form tags", zap.Strings("new_tags", newTags), zap.Any("form", c.Request.Form))

	// Get current org to compare tags
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Determine which tags to add and which to remove
	tagsToAdd, tagsToRemove := calculateTagChanges(org.Tags, newTags)

	// Call admin API to add tags if needed
	if len(tagsToAdd) > 0 {
		if err := s.callAdminAddTags(c, orgID, tagsToAdd); err != nil {
			s.l.Error("failed to add tags", zap.String("org_id", orgID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tags"})
			return
		}
	}

	// Call admin API to remove tags if needed
	if len(tagsToRemove) > 0 {
		if err := s.callAdminRemoveTags(c, orgID, tagsToRemove); err != nil {
			s.l.Error("failed to remove tags", zap.String("org_id", orgID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove tags"})
			return
		}
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

	// Call admin API to remove the tag
	if err := s.callAdminRemoveTags(c, orgID, []string{tagToRemove}); err != nil {
		s.l.Error("failed to remove tag", zap.String("org_id", orgID), zap.String("tag", tagToRemove), zap.Error(err))
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

// calculateTagChanges determines which tags need to be added and removed
func calculateTagChanges(currentTags, newTags []string) (toAdd, toRemove []string) {
	currentSet := make(map[string]bool)
	for _, tag := range currentTags {
		currentSet[tag] = true
	}

	newSet := make(map[string]bool)
	for _, tag := range newTags {
		newSet[tag] = true
	}

	// Find tags to add (in new but not in current)
	for _, tag := range newTags {
		if !currentSet[tag] {
			toAdd = append(toAdd, tag)
		}
	}

	// Find tags to remove (in current but not in new)
	for _, tag := range currentTags {
		if !newSet[tag] {
			toRemove = append(toRemove, tag)
		}
	}

	return toAdd, toRemove
}

// callAdminAddTags calls the admin API to add tags to an organization
func (s *service) callAdminAddTags(c *gin.Context, orgID string, tags []string) error {
	url := fmt.Sprintf("http://localhost:%s/v1/orgs/%s/admin-add-tags", s.cfg.InternalHTTPPort, orgID)

	requestBody := map[string]interface{}{
		"tags": tags,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Nuon-Admin-Email", "admin@nuon.co") // Admin dashboard has admin privileges

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("admin API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// callAdminRemoveTags calls the admin API to remove tags from an organization
func (s *service) callAdminRemoveTags(c *gin.Context, orgID string, tags []string) error {
	url := fmt.Sprintf("http://localhost:%s/v1/orgs/%s/admin-remove-tags", s.cfg.InternalHTTPPort, orgID)

	requestBody := map[string]interface{}{
		"tags": tags,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Nuon-Admin-Email", "admin@nuon.co") // Admin dashboard has admin privileges

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("admin API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
