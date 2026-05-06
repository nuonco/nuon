package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const orgsPerPage = 8

func (s *service) Orgs(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	label := c.Query("label")
	page := getPageFromQuery(c)

	orgs, totalPages, err := s.getOrgs(ctx, search, label, page)
	if err != nil {
		s.l.Error("failed to get orgs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	labelOptions, err := s.getOrgLabelOptions(ctx)
	if err != nil {
		s.l.Warn("failed to get label options", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"orgs":          orgs,
		"label_options": labelOptions,
		"page":          page,
		"total_pages":   totalPages,
	})
}

func (s *service) getOrgs(ctx context.Context, search, label string, page int) ([]*app.Org, int, error) {
	type OrgWithCounts struct {
		app.Org
		AppCount     int `gorm:"column:app_count"`
		InstallCount int `gorm:"column:install_count"`
	}

	var orgsWithCounts []OrgWithCounts
	var totalCount int64

	query := s.db.WithContext(ctx).Model(&app.Org{})

	if search != "" {
		search = strings.TrimSpace(search)
		query = query.Where(
			"name ILIKE ? OR id = ?",
			"%"+search+"%",
			search,
		)
	}

	if label != "" {
		if parts := strings.SplitN(label, ":", 2); len(parts) == 2 {
			query = query.Where("labels->>? = ?", parts[0], parts[1])
		} else {
			query = query.Where("labels::text ILIKE ?", "%"+label+"%")
		}
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count orgs: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(orgsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * orgsPerPage

	res := query.
		Select("orgs.*, " +
			"(SELECT COUNT(*) FROM apps WHERE apps.org_id = orgs.id AND apps.deleted_at = 0) as app_count, " +
			"(SELECT COUNT(*) FROM installs WHERE installs.org_id = orgs.id AND installs.deleted_at = 0) as install_count").
		Order("created_at desc").
		Limit(orgsPerPage).
		Offset(offset).
		Find(&orgsWithCounts)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get orgs: %w", res.Error)
	}

	orgs := make([]*app.Org, len(orgsWithCounts))
	for i := range orgsWithCounts {
		orgsWithCounts[i].Org.AppCount = orgsWithCounts[i].AppCount
		orgsWithCounts[i].Org.InstallCount = orgsWithCounts[i].InstallCount
		orgs[i] = &orgsWithCounts[i].Org
	}

	return orgs, totalPages, nil
}

type orgLabelOption struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func (s *service) getOrgLabelOptions(ctx context.Context) ([]orgLabelOption, error) {
	type row struct {
		Labels *string
	}
	var rows []row
	if err := s.db.WithContext(ctx).
		Table("orgs").
		Select("labels::text as labels").
		Where("deleted_at = 0").
		Where("labels IS NOT NULL").
		Where("labels::text != '{}'").
		Where("labels::text != 'null'").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	labelMap := make(map[string]map[string]bool)
	for _, r := range rows {
		if r.Labels == nil {
			continue
		}
		var parsed map[string]string
		if err := json.Unmarshal([]byte(*r.Labels), &parsed); err != nil {
			continue
		}
		for k, v := range parsed {
			if _, ok := labelMap[k]; !ok {
				labelMap[k] = make(map[string]bool)
			}
			labelMap[k][v] = true
		}
	}

	var options []orgLabelOption
	for k, vs := range labelMap {
		vals := make([]string, 0, len(vs))
		for v := range vs {
			vals = append(vals, v)
		}
		sort.Strings(vals)
		options = append(options, orgLabelOption{Key: k, Values: vals})
	}
	sort.Slice(options, func(i, j int) bool { return options[i].Key < options[j].Key })

	return options, nil
}
