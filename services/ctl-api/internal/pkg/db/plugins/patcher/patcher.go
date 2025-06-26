package patcher

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

func NewPatcherPlugin() *patcherPlugin {
	return &patcherPlugin{}
}

type patcherPlugin struct {
	models []interface{}
}

func (m *patcherPlugin) Name() string {
	return "patcher-plugin"
}

func (m *patcherPlugin) Initialize(db *gorm.DB) error {
	db.Callback().Update().Before("gorm:update").Register("enable_patcher_on_query", m.enablePatcher)
	return nil
}

func (m *patcherPlugin) enablePatcher(tx *gorm.DB) {
	enablePagination, ok := tx.Get(PatcherEnabledKey)
	if !(ok && enablePagination.(bool)) {
		return
	}

	ctxExclusions := []string{}
	ctxPatcher := cctx.PatcherFromContext(tx.Statement.Context)
	exclusions, ok := tx.Get(PatcherExclusionsKey)
	if !ok {
		exclusions = []string{}
	} else {
		ctxExclusions, _ = exclusions.([]string)
	}

	filteredProperties := filterProperties(ctxPatcher.SelectFields, ctxExclusions)

	tx.Select(filteredProperties)
}

// filterProperties removes exclusions from the properties slice
func filterProperties(properties []string, exclusions []string) []string {
	if len(exclusions) == 0 {
		return properties
	}

	// Create a map for fast lookup of exclusions
	excludeMap := make(map[string]bool, len(exclusions))
	for _, exclusion := range exclusions {
		excludeMap[exclusion] = true
	}

	// Filter properties
	var filtered []string
	for _, prop := range properties {
		if !excludeMap[prop] {
			filtered = append(filtered, prop)
		}
	}

	return filtered
}
