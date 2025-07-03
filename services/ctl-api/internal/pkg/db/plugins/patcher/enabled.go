package patcher

var (
	PatcherEnabledKey string = "patcher_enabled_key"
	PatcherOptionsKey string = "patcher_options_key"
)

type PatcherOptions struct {
	Exclusions []string
	Overrides  map[string]string
}
