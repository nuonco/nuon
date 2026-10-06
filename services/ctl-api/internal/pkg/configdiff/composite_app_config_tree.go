package configdiff

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	pkgdiff "github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ComputeCompositeAppConfigTree builds the hierarchical config diff the
// dashboard renders. Stack, runner, sandbox, and each component come from that
// entity's applied config. Every other section uses the install applied id.
// An empty baseline compares that slice to nothing.
func ComputeCompositeAppConfigTree(ctx context.Context, db *gorm.DB, baselines CompositeBaselines, newAppConfigID string) (*pkgdiff.Diff, error) {
	var newAppCfg app.AppConfig
	if err := preload(db.WithContext(ctx)).First(&newAppCfg, "id = ?", newAppConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to get new app config: %w", err)
	}

	loaded, err := loadBaselineConfigs(ctx, db, baselines, newAppConfigID, &newAppCfg)
	if err != nil {
		return nil, err
	}

	newIntermediate, ok, err := loadIntermediateConfig(ctx, &newAppCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to load new intermediate config: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("new app config %s has no intermediate config", newAppConfigID)
	}

	intermediates := map[string]*config.AppConfig{newAppConfigID: newIntermediate}
	for id, cfg := range loaded {
		if id == newAppConfigID {
			continue
		}
		intermediate, ok, err := loadIntermediateConfig(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("unable to load intermediate config %s: %w", id, err)
		}
		if ok {
			intermediates[id] = intermediate
		}
	}

	return spliceCompositeAppConfigTree(compositeTreeInputs{
		New:                newIntermediate,
		Intermediates:      intermediates,
		Baselines:          baselines,
		ComponentIDsByName: componentIDsByName(loaded),
	}), nil
}

func componentIDsByName(loaded map[string]*app.AppConfig) map[string]string {
	out := map[string]string{}
	for _, cfg := range loaded {
		if cfg == nil {
			continue
		}
		for _, conn := range cfg.ComponentConfigConnections {
			name := conn.ComponentName
			if name == "" {
				name = conn.Component.Name
			}
			if name == "" || conn.ComponentID == "" {
				continue
			}
			out[name] = conn.ComponentID
		}
	}
	return out
}

type compositeTreeInputs struct {
	New                *config.AppConfig
	Intermediates      map[string]*config.AppConfig
	Baselines          CompositeBaselines
	ComponentIDsByName map[string]string
	pairs              map[string]*pkgdiff.Diff
}

func spliceCompositeAppConfigTree(in compositeTreeInputs) *pkgdiff.Diff {
	if in.New == nil {
		return nil
	}
	in.pairs = map[string]*pkgdiff.Diff{}
	root := in.pair(in.Baselines.Fallback)

	if in.Baselines.Sandbox != in.Baselines.Fallback {
		replaceTopChild(root, "sandbox", topChild(in.pair(in.Baselines.Sandbox), "sandbox"))
	}
	if in.Baselines.Stack != in.Baselines.Fallback {
		stackPair := in.pair(in.Baselines.Stack)
		replaceTopChild(root, "stack", topChild(stackPair, "stack"))
		replaceTopChild(root, "runner", topChild(stackPair, "runner"))
	}

	in.spliceComponents(root)
	return root
}

func (in *compositeTreeInputs) pair(baselineID string) *pkgdiff.Diff {
	if tree, ok := in.pairs[baselineID]; ok {
		return tree
	}
	var old *config.AppConfig
	if baselineID != "" {
		old = in.Intermediates[baselineID]
	}
	tree := in.New.Diff(old)
	in.pairs[baselineID] = tree
	return tree
}

func (in *compositeTreeInputs) spliceComponents(root *pkgdiff.Diff) {
	names := map[string]struct{}{}
	collect := func(cfg *config.AppConfig) {
		if cfg == nil {
			return
		}
		for _, component := range cfg.Components {
			if component != nil && component.Name != "" {
				names[component.Name] = struct{}{}
			}
		}
	}
	collect(in.New)
	if in.Baselines.Fallback != "" {
		collect(in.Intermediates[in.Baselines.Fallback])
	}
	seenBaseline := map[string]struct{}{}
	for _, baselineID := range in.Baselines.Components {
		if baselineID == "" {
			continue
		}
		if _, ok := seenBaseline[baselineID]; ok {
			continue
		}
		seenBaseline[baselineID] = struct{}{}
		collect(in.Intermediates[baselineID])
	}

	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)

	for _, name := range ordered {
		baselineID := in.Baselines.Fallback
		explicit := ""
		if id := in.ComponentIDsByName[name]; id != "" {
			explicit = in.Baselines.Components[id]
		}
		if explicit != "" {
			baselineID = explicit
		}
		if baselineID == in.Baselines.Fallback {
			if componentNode(root, name) != nil {
				continue
			}
			node := in.componentOnOtherBaseline(name)
			if node == nil {
				continue
			}
			setComponent(root, name, node)
			continue
		}
		setComponent(root, name, componentNode(in.pair(baselineID), name))
	}
}

func (in *compositeTreeInputs) componentOnOtherBaseline(name string) *pkgdiff.Diff {
	seen := map[string]struct{}{}
	for _, baselineID := range in.Baselines.Components {
		if baselineID == "" || baselineID == in.Baselines.Fallback {
			continue
		}
		if _, ok := seen[baselineID]; ok {
			continue
		}
		seen[baselineID] = struct{}{}
		if node := componentNode(in.pair(baselineID), name); node != nil {
			return node
		}
	}
	return nil
}

func topChild(tree *pkgdiff.Diff, key string) *pkgdiff.Diff {
	if tree == nil {
		return nil
	}
	for _, child := range tree.Children {
		if child != nil && child.Key == key {
			return child
		}
	}
	return nil
}

func replaceTopChild(root *pkgdiff.Diff, key string, replacement *pkgdiff.Diff) {
	if root == nil {
		return
	}
	if replacement != nil {
		clearImpacts(replacement)
	}
	for i, child := range root.Children {
		if child != nil && child.Key == key {
			if replacement == nil {
				root.Children = append(root.Children[:i], root.Children[i+1:]...)
				return
			}
			root.Children[i] = replacement
			return
		}
	}
	if replacement != nil {
		root.Children = append(root.Children, replacement)
	}
}

func componentNode(tree *pkgdiff.Diff, name string) *pkgdiff.Diff {
	return topChild(topChild(tree, "components"), "component."+name)
}

func setComponent(root *pkgdiff.Diff, name string, node *pkgdiff.Diff) {
	if root == nil {
		return
	}
	section := topChild(root, "components")
	if section == nil {
		if node == nil {
			return
		}
		section = pkgdiff.NewDiff(pkgdiff.WithKey("components"))
		root.Children = append(root.Children, section)
	}
	key := "component." + name
	if node != nil {
		clearImpacts(node)
	}
	for i, child := range section.Children {
		if child != nil && child.Key == key {
			if node == nil {
				section.Children = append(section.Children[:i], section.Children[i+1:]...)
				return
			}
			section.Children[i] = node
			return
		}
	}
	if node != nil {
		section.Children = append(section.Children, node)
	}
}

func clearImpacts(node *pkgdiff.Diff) {
	if node == nil {
		return
	}
	node.Impacted = false
	node.ImpactReasons = nil
	for _, child := range node.Children {
		clearImpacts(child)
	}
}
