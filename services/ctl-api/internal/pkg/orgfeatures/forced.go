// Package orgfeatures holds deployment-wide org feature overrides. It is
// a leaf package so the config loader can populate it and GORM hooks, which
// have no access to the config object, can read it.
package orgfeatures

import (
	"strings"
	"sync/atomic"
)

var (
	forced atomic.Pointer[map[string]bool]
	auto   atomic.Pointer[map[string]bool]
)

func parseCSV(csv string) map[string]bool {
	set := make(map[string]bool)
	for _, name := range strings.Split(csv, ",") {
		if name = strings.TrimSpace(name); name != "" {
			set[name] = true
		}
	}
	return set
}

// SetForced parses the comma-separated forced_enabled_features config value.
func SetForced(csv string) {
	set := parseCSV(csv)
	forced.Store(&set)
}

// SetAuto parses the comma-separated auto_enabled_features config value.
// These are stored true on org create and can still be toggled off.
func SetAuto(csv string) {
	set := parseCSV(csv)
	auto.Store(&set)
}

// Forced returns the flags this deployment pins on for every org.
func Forced() map[string]bool {
	set := forced.Load()
	if set == nil {
		return map[string]bool{}
	}
	return *set
}

// IsForced reports whether the flag is pinned on for every org.
func IsForced(name string) bool {
	return Forced()[name]
}

// Auto returns flags this deployment turns on for newly created orgs.
func Auto() map[string]bool {
	set := auto.Load()
	if set == nil {
		return map[string]bool{}
	}
	return *set
}
