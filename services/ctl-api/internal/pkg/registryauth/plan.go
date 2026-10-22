package registryauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

const refreshTimeout = 30 * time.Second

const authKey = "OCIAuth"

type Refreshed struct {
	Path  []string
	Value any
}

type location struct {
	path []string
	get  func(*plantypes.CompositePlan) *configs.OCIRegistryRepository
}

var locations = []location{
	{[]string{"build_plan", "dst_registry"}, func(cp *plantypes.CompositePlan) *configs.OCIRegistryRepository {
		if cp.BuildPlan == nil {
			return nil
		}
		return cp.BuildPlan.Dst
	}},
	{[]string{"build_plan", "container_image_pull_plan", "repo_config"}, func(cp *plantypes.CompositePlan) *configs.OCIRegistryRepository {
		if cp.BuildPlan == nil || cp.BuildPlan.ContainerImagePullPlan == nil {
			return nil
		}
		return cp.BuildPlan.ContainerImagePullPlan.RepoCfg
	}},
	{[]string{"deploy_plan", "src_registry"}, func(cp *plantypes.CompositePlan) *configs.OCIRegistryRepository {
		if cp.DeployPlan == nil {
			return nil
		}
		return cp.DeployPlan.Src
	}},
	{[]string{"action_workflow_run_plan", "image_registry"}, func(cp *plantypes.CompositePlan) *configs.OCIRegistryRepository {
		if cp.ActionWorkflowRunPlan == nil {
			return nil
		}
		return cp.ActionWorkflowRunPlan.ImageRegistry
	}},
	{[]string{"sync_oci_plan", "src_registry"}, func(cp *plantypes.CompositePlan) *configs.OCIRegistryRepository {
		if cp.SyncOCIPlan == nil {
			return nil
		}
		return cp.SyncOCIPlan.Src
	}},
	{[]string{"sync_oci_plan", "dst_registry"}, func(cp *plantypes.CompositePlan) *configs.OCIRegistryRepository {
		if cp.SyncOCIPlan == nil {
			return nil
		}
		return cp.SyncOCIPlan.Dst
	}},
	{[]string{"fetch_image_metadata_plan", "registry"}, func(cp *plantypes.CompositePlan) *configs.OCIRegistryRepository {
		if cp.FetchImageMetadataPlan == nil {
			return nil
		}
		return cp.FetchImageMetadataPlan.Registry
	}},
	{[]string{"sandbox_run_plan", "oci_source", "registry"}, func(cp *plantypes.CompositePlan) *configs.OCIRegistryRepository {
		if cp.SandboxRunPlan == nil || cp.SandboxRunPlan.OCISource == nil {
			return nil
		}
		return cp.SandboxRunPlan.OCISource.Registry
	}},
}

func (i *Issuer) RefreshCompositePlan(ctx context.Context, l *zap.Logger, orgID string, cp *plantypes.CompositePlan) ([]Refreshed, error) {
	if cp == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()

	minted := map[string]*configs.OCIRegistryAuth{}
	var refreshed []Refreshed
	for _, loc := range locations {
		cfg := loc.get(cp)
		src := i.refreshSource(orgID, cfg)
		if src == nil {
			continue
		}

		auth, ok := minted[src.key]
		if !ok {
			var err error
			auth, err = src.mint(ctx, l)
			if err != nil {
				return nil, fmt.Errorf("unable to refresh registry credentials at %s: %w", strings.Join(loc.path, "."), err)
			}
			minted[src.key] = auth
		}

		cfg.OCIAuth = &configs.OCIRegistryAuth{Username: auth.Username, Password: auth.Password}
		path := append(append([]string{}, loc.path...), authKey)
		refreshed = append(refreshed, Refreshed{Path: path, Value: *cfg.OCIAuth})
	}

	return refreshed, nil
}

func PatchSubPlanJSON(raw []byte, refreshed []Refreshed) ([]byte, error) {
	out := json.RawMessage(raw)
	for _, r := range refreshed {
		value, err := marshal(r.Value)
		if err != nil {
			return nil, err
		}

		out, err = setJSONPath(out, r.Path[1:], value)
		if err != nil {
			return nil, fmt.Errorf("unable to patch %s: %w", strings.Join(r.Path[1:], "."), err)
		}
	}

	return out, nil
}

func setJSONPath(raw json.RawMessage, path []string, value json.RawMessage) (json.RawMessage, error) {
	if len(path) == 0 {
		return value, nil
	}
	if bytes.HasPrefix(bytes.TrimSpace(raw), []byte("[")) {
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, err
		}
		idx, err := strconv.Atoi(path[0])
		if err != nil || idx < 0 || idx >= len(items) {
			return nil, fmt.Errorf("invalid array index %s", path[0])
		}
		items[idx], err = setJSONPath(items[idx], path[1:], value)
		if err != nil {
			return nil, err
		}
		return marshal(items)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, fmt.Errorf("%s is not an object", path[0])
	}

	if len(path) == 1 {
		obj[path[0]] = value
		return marshal(obj)
	}

	child, ok := obj[path[0]]
	if !ok {
		return nil, fmt.Errorf("missing %s", path[0])
	}
	updated, err := setJSONPath(child, path[1:], value)
	if err != nil {
		return nil, err
	}
	obj[path[0]] = updated

	return marshal(obj)
}

func marshal(v any) (json.RawMessage, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
