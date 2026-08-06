package airgap

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/render"
	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
	statepkg "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// VirtualInstallID is the synthetic install ID CompilePlanEnvelope freezes
// into a zero-install bundle. Root stack template compilation must use the
// same ID: nuon-bundle's deployment-time install-ID substitution rewrites
// only occurrences of the envelope's frozen ID, so a template compiled under
// any other ID would keep its physical IAM/log/secret names constant across
// deployments of the same bundle and collide within one account.
func VirtualInstallID(appConfigID string) string {
	return virtualID("vinst", appConfigID)
}

// RenderConfigForStackCompile mirrors the connected install flow
// (generateinstallstackversion), which renders these config blocks against
// install state before generating the stack template. Zero-install
// compilation has no install state, so templates render against the
// synthetic install identity; install inputs are seeded as publish-time
// placeholder tokens, which root template validation rejects if they leak
// into the compiled stack. Returns a copy: rendered blocks are deep-copied
// so the caller's config is never mutated.
func RenderConfigForStackCompile(ctx context.Context, db *gorm.DB, cfg *app.AppConfig, installID string) (*app.AppConfig, error) {
	inputs, err := exportInputSpecs(ctx, db, cfg.ID)
	if err != nil {
		return nil, err
	}
	return renderConfigForStackCompile(cfg, installID, inputs)
}

func renderConfigForStackCompile(cfg *app.AppConfig, installID string, inputs []runnerairgap.InputSpec) (*app.AppConfig, error) {
	inputValues := make(map[string]string, len(inputs))
	for _, input := range inputs {
		inputValues[input.Name] = runnerairgap.InputPlaceholder(input.Name)
	}
	st := &statepkg.State{
		ID:      installID,
		Name:    "airgap",
		Org:     &statepkg.OrgState{ID: cfg.OrgID, Populated: true},
		App:     &statepkg.AppState{ID: cfg.AppID, Populated: true, Variables: map[string]string{}},
		Install: &statepkg.InstallState{ID: installID, Name: "airgap", Populated: true, Inputs: inputValues},
		Inputs:  &statepkg.InputsState{Populated: true, Inputs: inputValues},
		Cloud:   compileCloudAccount(cfg.StackConfig.Type),
	}
	inner, err := st.AsMap()
	if err != nil {
		return nil, fmt.Errorf("encode compile state: %w", err)
	}
	stateData := map[string]any{"nuon": inner}

	out := *cfg
	if err := deepCopyConfigBlock(&cfg.PermissionsConfig, &out.PermissionsConfig); err != nil {
		return nil, fmt.Errorf("copy permissions config: %w", err)
	}
	if err := deepCopyConfigBlock(&cfg.BreakGlassConfig, &out.BreakGlassConfig); err != nil {
		return nil, fmt.Errorf("copy break glass config: %w", err)
	}
	if err := deepCopyConfigBlock(&cfg.SecretsConfig, &out.SecretsConfig); err != nil {
		return nil, fmt.Errorf("copy secrets config: %w", err)
	}
	if err := deepCopyConfigBlock(&cfg.StackConfig, &out.StackConfig); err != nil {
		return nil, fmt.Errorf("copy stack config: %w", err)
	}

	if err := render.RenderStruct(&out.PermissionsConfig, stateData); err != nil {
		return nil, fmt.Errorf("render permissions config: %w", err)
	}
	if err := render.RenderStruct(&out.BreakGlassConfig, stateData); err != nil {
		return nil, fmt.Errorf("render break glass config: %w", err)
	}
	if err := render.RenderStruct(&out.SecretsConfig, stateData); err != nil {
		return nil, fmt.Errorf("render secrets config: %w", err)
	}
	if err := render.RenderStruct(&out.StackConfig, stateData); err != nil {
		return nil, fmt.Errorf("render stack config: %w", err)
	}
	for i := range out.SecretsConfig.Secrets {
		out.SecretsConfig.Secrets[i].UpdateCloudformationStackInfo()
	}
	if err := config.RenderCustomNestedStackParameters(out.StackConfig.CustomNestedStacks, stateData); err != nil {
		return nil, fmt.Errorf("render custom nested stack parameters: %w", err)
	}
	return &out, nil
}

// Only the audit/relation fields are dropped by json:"-", none of which feed
// template generation.
func deepCopyConfigBlock[T any](src, dst *T) error {
	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}
	var fresh T
	if err := json.Unmarshal(raw, &fresh); err != nil {
		return err
	}
	*dst = fresh
	return nil
}
