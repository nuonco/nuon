package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/dominikbraun/graph/draw"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const orgInstallsPerPage = 8

func (s *service) OrgDetail(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Org ID is required"})
		return
	}

	page := getPageFromQuery(c)

	// Fetch data in parallel
	var (
		org                         *app.Org
		installs                    []*app.Install
		installsTotalPages          int
		recentApp                   *app.App
		orgErr, installsErr, appErr error
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		org, orgErr = s.getOrg(gCtx, orgID)
		return orgErr
	})

	g.Go(func() error {
		installs, installsTotalPages, installsErr = s.getInstallsForOrg(gCtx, orgID, page)
		return installsErr
	})

	g.Go(func() error {
		recentApp, appErr = s.getMostRecentApp(gCtx, orgID)
		return appErr
	})

	if err := g.Wait(); err != nil {
		s.l.Error("failed to fetch data", zap.String("org_id", orgID), zap.Error(err))
		if orgErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		}
		return
	}

	// Fetch graph after getting the app (only if app exists)
	var graphDot string
	if recentApp != nil {
		var err error
		graphDot, err = s.getAppComponentGraph(ctx, recentApp.ID)
		if err != nil {
			// Log but don't fail the page if graph fetch fails
			s.l.Warn("failed to fetch component graph", zap.String("app_id", recentApp.ID), zap.Error(err))
		}
	}

	component := views.OrgDetail(org, installs, recentApp, graphDot, s.cfg.AppURL, page, installsTotalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getOrg(ctx context.Context, orgID string) (*app.Org, error) {
	var org app.Org

	res := s.db.WithContext(ctx).
		Where("id = ?", orgID).
		First(&org)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	return &org, nil
}

func (s *service) getInstallsForOrg(ctx context.Context, orgID string, page int) ([]*app.Install, int, error) {
	var installs []*app.Install
	var totalCount int64

	// Build base query
	query := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Where("org_id = ?", orgID)

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count installs: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(orgInstallsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Calculate offset
	offset := (page - 1) * orgInstallsPerPage

	// Get paginated results
	res := query.
		Preload("App").
		Preload("RunnerGroup.Runners").
		Preload("AppConfig").
		Preload("AppRunnerConfig").
		Order("created_at desc").
		Limit(orgInstallsPerPage).
		Offset(offset).
		Find(&installs)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	return installs, totalPages, nil
}

func (s *service) getMostRecentApp(ctx context.Context, orgID string) (*app.App, error) {
	var app app.App

	res := s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Order("updated_at DESC").
		First(&app)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No apps found, not an error
		}
		return nil, fmt.Errorf("unable to get most recent app: %w", res.Error)
	}

	return &app, nil
}

func (s *service) getAppComponentGraph(ctx context.Context, appID string) (string, error) {
	// 1. Get the latest app config
	appConfig, err := s.appsHelpers.GetAppLatestConfig(ctx, appID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // No config yet, not an error
		}
		return "", fmt.Errorf("unable to get latest app config: %w", err)
	}

	// 2. Get the full app config with all component connections
	fullConfig, err := s.appsHelpers.GetFullAppConfig(ctx, appConfig.ID, true)
	if err != nil {
		return "", fmt.Errorf("unable to get full app config: %w", err)
	}

	// 3. Generate the component graph
	graph, err := s.appsHelpers.GetConfigGraph(ctx, fullConfig)
	if err != nil {
		return "", fmt.Errorf("unable to generate config graph: %w", err)
	}

	// 4. Render graph to DOT format with styling attributes
	var buf bytes.Buffer
	if err := draw.DOT(graph, &buf,
		draw.GraphAttribute("name", "name"),
		draw.GraphAttribute("rankdir", "LR"),
		draw.GraphAttribute("bgcolor", "transparent"),
		draw.GraphAttribute("nodesep", "0.8"),
		draw.GraphAttribute("ranksep", "1.2"),
		draw.GraphAttribute("splines", "ortho"),
	); err != nil {
		return "", fmt.Errorf("unable to render graph: %w", err)
	}

	// Post-process DOT to add node styling and wrap long labels
	dotString := buf.String()
	dotString = addNodeStyling(dotString)
	dotString = wrapLongLabels(dotString)
	return dotString, nil
}

func addNodeStyling(dotString string) string {
	// Insert node styling after the first opening brace
	insertPos := bytes.Index([]byte(dotString), []byte("{\n"))
	if insertPos == -1 {
		return dotString
	}

	// Position after "{\n"
	insertPos += 2

	// Build the new DOT string with node styling inserted
	result := dotString[:insertPos] +
		"    node [shape=box, style=filled, width=1.5, height=0.6, fixedsize=false, margin=0.2];\n" +
		dotString[insertPos:]

	return result
}

func wrapLongLabels(dotString string) string {
	// Find and wrap long labels by splitting on underscores
	// Pattern: label="some_long_name"
	lines := bytes.Split([]byte(dotString), []byte("\n"))

	for i, line := range lines {
		// Look for label= attributes
		if bytes.Contains(line, []byte("label=")) {
			// Extract the label value
			start := bytes.Index(line, []byte("label=\""))
			if start == -1 {
				continue
			}
			start += 7 // Move past 'label="'

			end := bytes.Index(line[start:], []byte("\""))
			if end == -1 {
				continue
			}

			label := string(line[start : start+end])

			// If label is long (more than 15 chars), split on underscores
			if len(label) > 15 {
				wrapped := wrapLabel(label)
				// Replace the label in the line
				newLine := bytes.Replace(line, []byte("label=\""+label+"\""), []byte("label=\""+wrapped+"\""), 1)
				lines[i] = newLine
			}
		}
	}

	return string(bytes.Join(lines, []byte("\n")))
}

func wrapLabel(label string) string {
	// For very long labels (> 15 chars), split on underscores
	if len(label) <= 15 {
		return label
	}

	// Split on underscores
	parts := strings.Split(label, "_")
	if len(parts) == 1 {
		return label // No underscores, can't wrap nicely
	}

	// Build lines, trying to keep each line under 15 characters
	var lines []string
	currentLine := ""

	for i, part := range parts {
		testLine := currentLine
		if testLine != "" {
			testLine += "_"
		}
		testLine += part

		// If this would make the line too long, start a new line
		if len(testLine) > 15 && currentLine != "" {
			lines = append(lines, currentLine)
			currentLine = part
		} else {
			currentLine = testLine
		}

		// Add the last line
		if i == len(parts)-1 && currentLine != "" {
			lines = append(lines, currentLine)
		}
	}

	// Join lines with \n for Graphviz line breaks
	return strings.Join(lines, "\\n")
}
