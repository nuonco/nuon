package installs

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/config/refs"
)

// stackReprovisionWarning reports changed inputs that the install stack renders.
// Those values are stored by this command, and the running stack keeps the
// previous ones until it is reprovisioned. A failure to load the app config
// leaves the update in place and skips the warning.
func (s *Service) stackReprovisionWarning(ctx context.Context, install *models.AppInstall, changed map[string]bool) string {
	if install == nil || install.AppConfigID == "" || len(changed) == 0 {
		return ""
	}

	recurse := true
	appCfg, err := s.api.GetAppConfig(ctx, install.AppID, install.AppConfigID, &recurse)
	if err != nil {
		ui.PrintDebug(fmt.Sprintf("unable to check stack input references: %v", err))
		return ""
	}

	var installCfg *models.AppInstallConfig
	if install.InstallConfig != nil {
		installCfg = install.InstallConfig
	}
	return formatStackReprovisionWarning(changedStackInputs(changed, inputNamesReferencedByStack(appCfg, installCfg)))
}

// inputNamesReferencedByStack returns install input names referenced by the
// configs rendered when an install stack version is generated: the stack
// itself, permissions, break-glass, secrets, and per-install stack overrides.
func inputNamesReferencedByStack(appCfg *models.AppAppConfig, installCfg *models.AppInstallConfig) []string {
	names := map[string]struct{}{}
	add := func(obj any) {
		for _, name := range inputRefNames(obj) {
			names[name] = struct{}{}
		}
	}
	if appCfg != nil {
		add(appCfg.Stack)
		add(appCfg.Permissions)
		add(appCfg.BreakGlass)
		add(appCfg.Secrets)
	}
	add(installCfg)

	out := make([]string, 0, len(names))
	for name := range names {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func inputRefNames(obj any) []string {
	if isNil(obj) {
		return nil
	}
	parsed, err := refs.Parse(obj)
	if err != nil {
		return nil
	}
	var names []string
	for _, ref := range parsed {
		if ref.Type == refs.RefTypeInputs || ref.Type == refs.RefTypeInstallInputs {
			names = append(names, ref.Name)
		}
	}
	return names
}

func changedStackInputs(changed map[string]bool, referenced []string) []string {
	var out []string
	for _, name := range referenced {
		if changed[name] {
			out = append(out, name)
		}
	}
	return out
}

func formatStackReprovisionWarning(names []string) string {
	if len(names) == 0 {
		return ""
	}
	list := names[0]
	these := "this change"
	switch len(names) {
	case 2:
		list = names[0] + " and " + names[1]
		these = "these changes"
	default:
		if len(names) > 2 {
			list = strings.Join(names[:len(names)-1], ", ") + ", and " + names[len(names)-1]
			these = "these changes"
		}
	}
	return fmt.Sprintf("The install stack uses %s. Run nuon installs stacks reprovision to apply %s to the stack", list, these)
}

func isNil(obj any) bool {
	if obj == nil {
		return true
	}
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}
