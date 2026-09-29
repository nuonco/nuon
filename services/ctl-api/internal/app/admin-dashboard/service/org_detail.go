package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/dominikbraun/graph/draw"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

	var graphDot string
	if recentApp != nil {
		var err error
		graphDot, err = s.getAppComponentGraph(ctx, recentApp.ID)
		if err != nil {
			s.l.Warn("failed to fetch component graph", zap.String("app_id", recentApp.ID), zap.Error(err))
		}
	}

	storedFeatures, err := s.getStoredOrgFeatures(ctx, orgID)
	if err != nil {
		s.l.Warn("failed to fetch stored org features", zap.String("org_id", orgID), zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"org":                  org,
		"stored_features":      storedFeatures,
		"installs":             installs,
		"recent_app":           recentApp,
		"graph_dot":            graphDot,
		"app_url":              s.cfg.AppURL,
		"page":                 page,
		"installs_total_pages": installsTotalPages,
	})
}

func (s *service) getStoredOrgFeatures(ctx context.Context, orgID string) (map[string]bool, error) {
	var raw *string
	if err := s.readDB().WithContext(ctx).
		Raw("SELECT features FROM orgs WHERE id = ?", orgID).
		Scan(&raw).Error; err != nil {
		return nil, fmt.Errorf("unable to read org features: %w", err)
	}

	stored := make(map[string]bool)
	if raw == nil {
		return stored, nil
	}
	if err := json.Unmarshal([]byte(*raw), &stored); err != nil {
		return map[string]bool{}, nil
	}
	return stored, nil
}

func (s *service) getOrg(ctx context.Context, orgID string) (*app.Org, error) {
	var org app.Org

	res := s.readDB().WithContext(ctx).
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

	query := s.readDB().WithContext(ctx).
		Model(&app.Install{}).
		Unscoped().
		Where("org_id = ?", orgID)

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count installs: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(orgInstallsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * orgInstallsPerPage

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

	res := s.readDB().WithContext(ctx).
		Where("org_id = ?", orgID).
		Order("updated_at DESC").
		First(&app)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get most recent app: %w", res.Error)
	}

	return &app, nil
}

func (s *service) getAppComponentGraph(ctx context.Context, appID string) (string, error) {
	appConfig, err := s.appsHelpers.GetLatestActiveAppConfig(ctx, appID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("unable to get latest app config: %w", err)
	}

	fullConfig, err := s.appsHelpers.GetFullAppConfig(ctx, appConfig.ID, true)
	if err != nil {
		return "", fmt.Errorf("unable to get full app config: %w", err)
	}

	graph, err := s.appsHelpers.GetConfigGraph(ctx, fullConfig)
	if err != nil {
		return "", fmt.Errorf("unable to generate config graph: %w", err)
	}

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

	dotString := buf.String()
	dotString = addNodeStyling(dotString)
	dotString = wrapLongLabels(dotString)
	return dotString, nil
}

func addNodeStyling(dotString string) string {
	insertPos := bytes.Index([]byte(dotString), []byte("{\n"))
	if insertPos == -1 {
		return dotString
	}

	insertPos += 2

	result := dotString[:insertPos] +
		"    node [shape=box, style=filled, width=1.5, height=0.6, fixedsize=false, margin=0.2];\n" +
		dotString[insertPos:]

	return result
}

func wrapLongLabels(dotString string) string {
	lines := bytes.Split([]byte(dotString), []byte("\n"))

	for i, line := range lines {
		if bytes.Contains(line, []byte("label=")) {
			start := bytes.Index(line, []byte("label=\""))
			if start == -1 {
				continue
			}
			start += 7

			end := bytes.Index(line[start:], []byte("\""))
			if end == -1 {
				continue
			}

			label := string(line[start : start+end])

			if len(label) > 15 {
				wrapped := wrapLabel(label)
				newLine := bytes.Replace(line, []byte("label=\""+label+"\""), []byte("label=\""+wrapped+"\""), 1)
				lines[i] = newLine
			}
		}
	}

	return string(bytes.Join(lines, []byte("\n")))
}

func wrapLabel(label string) string {
	if len(label) <= 15 {
		return label
	}

	parts := strings.Split(label, "_")
	if len(parts) == 1 {
		return label
	}

	var lines []string
	currentLine := ""

	for i, part := range parts {
		testLine := currentLine
		if testLine != "" {
			testLine += "_"
		}
		testLine += part

		if len(testLine) > 15 && currentLine != "" {
			lines = append(lines, currentLine)
			currentLine = part
		} else {
			currentLine = testLine
		}

		if i == len(parts)-1 && currentLine != "" {
			lines = append(lines, currentLine)
		}
	}

	return strings.Join(lines, "\\n")
}
