package v2

import (
	"strings"
	"testing"

	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownsyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog/allsignals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// These properties pin the retry contract for apply signals: a retried apply
// must always be preceded by a fresh, approval-gated plan of the paired type
// targeting the same entity. Both retry entry points depend on this contract:
//
//   - executeworkflowstep/process_errors.go (auto-retry) and
//     executeworkflowstep/update_create_step_retry.go (manual retry) write a
//     retry-group directive when the signal's RetryGroup() is true, cloning
//     the whole plan+apply group.
//   - executeworkflowstep/clone.go (single-step clone fallback) expands the
//     apply via Clone() into [plan, apply].
//
// If either property breaks, an apply can re-run against a stale or
// unapproved plan.

const applyPlanTypeSuffix = "-apply-plan"

type planApplyIdentity struct {
	InstallID   string
	TargetID    string
	ComponentID string
}

type applyPlanPair struct {
	applyType signal.SignalType
	planType  signal.SignalType
	make      func(id planApplyIdentity, flowID string, sandboxMode bool) signal.Signal
}

var applyPlanPairs = []applyPlanPair{
	{
		applyType: componentdeployapplyplan.SignalType,
		planType:  componentdeploysyncandplan.SignalType,
		make: func(id planApplyIdentity, flowID string, sandboxMode bool) signal.Signal {
			return &componentdeployapplyplan.Signal{
				InstallComponentID: id.TargetID,
				InstallID:          id.InstallID,
				ComponentID:        id.ComponentID,
				FlowID:             flowID,
				SandboxMode:        sandboxMode,
			}
		},
	},
	{
		applyType: componentteardownapplyplan.SignalType,
		planType:  componentteardownsyncandplan.SignalType,
		make: func(id planApplyIdentity, flowID string, sandboxMode bool) signal.Signal {
			return &componentteardownapplyplan.Signal{
				InstallComponentID: id.TargetID,
				InstallID:          id.InstallID,
				ComponentID:        id.ComponentID,
				FlowID:             flowID,
				SandboxMode:        sandboxMode,
			}
		},
	},
	{
		applyType: provisionsandboxapplyplan.SignalType,
		planType:  provisionsandboxplan.SignalType,
		make: func(id planApplyIdentity, flowID string, sandboxMode bool) signal.Signal {
			return &provisionsandboxapplyplan.Signal{
				InstallSandboxID: id.TargetID,
				InstallID:        id.InstallID,
				FlowID:           flowID,
				SandboxMode:      sandboxMode,
			}
		},
	},
	{
		applyType: deprovisionsandboxapplyplan.SignalType,
		planType:  deprovisionsandboxplan.SignalType,
		make: func(id planApplyIdentity, flowID string, sandboxMode bool) signal.Signal {
			return &deprovisionsandboxapplyplan.Signal{
				InstallSandboxID: id.TargetID,
				InstallID:        id.InstallID,
				FlowID:           flowID,
				SandboxMode:      sandboxMode,
			}
		},
	},
	{
		applyType: reprovisionsandboxapplyplan.SignalType,
		planType:  reprovisionsandboxplan.SignalType,
		make: func(id planApplyIdentity, flowID string, sandboxMode bool) signal.Signal {
			return &reprovisionsandboxapplyplan.Signal{
				InstallSandboxID: id.TargetID,
				InstallID:        id.InstallID,
				FlowID:           flowID,
				SandboxMode:      sandboxMode,
			}
		},
	},
}

func planApplyIdentityOf(sig signal.Signal) (planApplyIdentity, bool) {
	switch s := sig.(type) {
	case *componentdeploysyncandplan.Signal:
		return planApplyIdentity{s.InstallID, s.InstallComponentID, s.ComponentID}, true
	case *componentdeployapplyplan.Signal:
		return planApplyIdentity{s.InstallID, s.InstallComponentID, s.ComponentID}, true
	case *componentteardownsyncandplan.Signal:
		return planApplyIdentity{s.InstallID, s.InstallComponentID, s.ComponentID}, true
	case *componentteardownapplyplan.Signal:
		return planApplyIdentity{s.InstallID, s.InstallComponentID, s.ComponentID}, true
	case *provisionsandboxplan.Signal:
		return planApplyIdentity{InstallID: s.InstallID, TargetID: s.InstallSandboxID}, true
	case *provisionsandboxapplyplan.Signal:
		return planApplyIdentity{InstallID: s.InstallID, TargetID: s.InstallSandboxID}, true
	case *deprovisionsandboxplan.Signal:
		return planApplyIdentity{InstallID: s.InstallID, TargetID: s.InstallSandboxID}, true
	case *deprovisionsandboxapplyplan.Signal:
		return planApplyIdentity{InstallID: s.InstallID, TargetID: s.InstallSandboxID}, true
	case *reprovisionsandboxplan.Signal:
		return planApplyIdentity{InstallID: s.InstallID, TargetID: s.InstallSandboxID}, true
	case *reprovisionsandboxapplyplan.Signal:
		return planApplyIdentity{InstallID: s.InstallID, TargetID: s.InstallSandboxID}, true
	}
	return planApplyIdentity{}, false
}

// TestApplySignalPairTableIsComplete guards the pair table itself: any
// registered signal whose type ends in "-apply-plan" must be covered by
// applyPlanPairs, so a future apply signal cannot silently skip these
// properties.
func TestApplySignalPairTableIsComplete(t *testing.T) {
	covered := make(map[signal.SignalType]bool, len(applyPlanPairs))
	for _, p := range applyPlanPairs {
		covered[p.applyType] = true
	}
	for typ := range catalog.SignalCatalog {
		if !strings.HasSuffix(string(typ), applyPlanTypeSuffix) {
			continue
		}
		if !covered[typ] {
			t.Errorf("apply signal %q is registered but not covered by applyPlanPairs; add it so the plan+apply retry properties cover it", typ)
		}
	}
}

// TestApplySignalsAlwaysRetryAsGroup asserts every apply signal opts into
// group retry, so both auto-retry and manual retry re-run the plan alongside
// the apply.
func TestApplySignalsAlwaysRetryAsGroup(t *testing.T) {
	for _, pair := range applyPlanPairs {
		sig := pair.make(planApplyIdentity{InstallID: "inst", TargetID: "target", ComponentID: "comp"}, "flow", false)
		rg, ok := sig.(signal.SignalWithRetryGroup)
		if !ok {
			t.Errorf("%s does not implement SignalWithRetryGroup: retrying it would clone the apply without re-running its plan group", pair.applyType)
			continue
		}
		if !rg.RetryGroup() {
			t.Errorf("%s RetryGroup() = false: retrying it would clone the apply without re-running its plan group", pair.applyType)
		}
	}
}

// TestApplyRetryAlwaysReplansProperty asserts the single-step clone fallback:
// for any apply signal with any target identity, Clone() must produce an
// approval-gated plan of the paired type before the apply, both preserving
// the original target identity.
func TestApplyRetryAlwaysReplansProperty(t *testing.T) {
	t.Run("clone_emits_plan_before_apply", hegel.Case(func(ht *hegel.T) {
		pair := hegel.Draw(ht, hegel.SampledFrom(applyPlanPairs))
		id := planApplyIdentity{
			InstallID:   hegel.Draw(ht, hegel.Text(1, 30)),
			TargetID:    hegel.Draw(ht, hegel.Text(1, 30)),
			ComponentID: hegel.Draw(ht, hegel.Text(1, 30)),
		}
		flowID := hegel.Draw(ht, hegel.Text(0, 30))
		sandboxMode := hegel.Draw(ht, hegel.Booleans())
		stepName := hegel.Draw(ht, hegel.Text(1, 40))

		applySig := pair.make(id, flowID, sandboxMode)
		wantID, ok := planApplyIdentityOf(applySig)
		if !ok {
			ht.Fatalf("%s missing from planApplyIdentityOf", pair.applyType)
		}

		cl, ok := applySig.(signal.SignalWithCloneSteps)
		if !ok {
			ht.Fatalf("%s does not implement SignalWithCloneSteps", pair.applyType)
		}
		defs, err := cl.Clone(nil, stepName)
		if err != nil {
			ht.Fatalf("%s Clone failed: %v", pair.applyType, err)
		}

		planIdx, applyIdx := -1, -1
		for i, def := range defs {
			switch def.Signal.Type() {
			case pair.planType:
				if planIdx != -1 {
					ht.Fatalf("%s Clone emitted more than one plan step", pair.applyType)
				}
				planIdx = i
			case pair.applyType:
				if applyIdx != -1 {
					ht.Fatalf("%s Clone emitted more than one apply step", pair.applyType)
				}
				applyIdx = i
			}
		}
		if planIdx == -1 {
			ht.Fatalf("%s Clone emitted no %s plan step: a retried apply would run without a fresh plan", pair.applyType, pair.planType)
		}
		if applyIdx == -1 {
			ht.Fatalf("%s Clone emitted no apply step", pair.applyType)
		}
		if planIdx >= applyIdx {
			ht.Fatalf("%s Clone ordered apply (idx %d) before plan (idx %d)", pair.applyType, applyIdx, planIdx)
		}
		if defs[planIdx].ExecutionType != "approval" {
			ht.Fatalf("%s cloned plan has execution type %q, want \"approval\": the re-plan would skip the approval gate", pair.applyType, defs[planIdx].ExecutionType)
		}

		for _, i := range []int{planIdx, applyIdx} {
			gotID, ok := planApplyIdentityOf(defs[i].Signal)
			if !ok {
				ht.Fatalf("%s clone def %d has unrecognized signal type %s", pair.applyType, i, defs[i].Signal.Type())
			}
			if gotID != wantID {
				ht.Fatalf("%s clone def %d changed target identity: got %+v, want %+v", pair.applyType, i, gotID, wantID)
			}
		}
	}, hegel.WithTestCases(200)))
}
