package service

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

var (
	installHealthFilterValues         = []string{"unhealthy", "degraded", "other", "unknown", "progressing", "healthy", "unset"}
	installLifecyclePhaseFilterValues = []string{"provisioning", "provisioned", "deprovisioning", "deprovisioned", "reprovisioning", "unset", "other"}
)

type installListFilters struct {
	health          []string
	lifecyclePhases []string
}

func parseInstallListFilters(ctx *gin.Context) (installListFilters, error) {
	health, err := parseInstallListFilter(ctx, "health", installHealthFilterValues)
	if err != nil {
		return installListFilters{}, err
	}
	lifecyclePhases, err := parseInstallListFilter(ctx, "lifecycle_phase", installLifecyclePhaseFilterValues)
	if err != nil {
		return installListFilters{}, err
	}
	return installListFilters{health: health, lifecyclePhases: lifecyclePhases}, nil
}

func parseInstallListFilter(ctx *gin.Context, name string, allowed []string) ([]string, error) {
	var values []string
	for _, raw := range ctx.QueryArray(name) {
		for _, part := range strings.Split(raw, ",") {
			value := strings.TrimSpace(part)
			if value == "" || slices.Contains(values, value) {
				continue
			}
			if !slices.Contains(allowed, value) {
				return nil, stderr.ErrUser{
					Err:         fmt.Errorf("unsupported %s filter value %q", name, value),
					Description: fmt.Sprintf("%s must be a comma-separated list of: %s", name, strings.Join(allowed, ", ")),
				}
			}
			values = append(values, value)
		}
	}
	return values, nil
}

func (f installListFilters) apply(tx *gorm.DB, table string) *gorm.DB {
	if len(f.health) == 0 && len(f.lifecyclePhases) == 0 {
		return tx
	}
	if len(f.health) > 0 {
		tx = tx.Where(viewsql.InstallHealthRollupV1(table+".id")+" IN ?", f.health)
	}
	if len(f.lifecyclePhases) > 0 {
		tx = tx.Where(viewsql.InstallLifecyclePhaseV1(table+".lifecycle_phase")+" IN ?", f.lifecyclePhases)
	}
	return tx.Order(table + ".id ASC")
}
