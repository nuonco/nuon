package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

// why: Subscribe-modal Block-Kit identifiers. Centralised so the open / re-render
// / view_submission handlers all key off the same ids — Slack rejects
// duplicate ids and silently drops mismatches when reading view.state.values.
const (
	subscribeModalCallbackID = "nuon_subscribe_modal"

	subscribeOrgBlockID  = "nuon_subscribe_org_block"
	subscribeOrgActionID = "nuon_subscribe_org"

	subscribeMatchBlockID  = "nuon_subscribe_match_block"
	subscribeMatchActionID = "nuon_subscribe_match"
	matchOptionAll         = "all"
	matchOptionSpecific    = "specific"

	subscribeKindBlockID  = "nuon_subscribe_kind_block"
	subscribeKindActionID = "nuon_subscribe_kind"
	kindOptionInstalls    = "installs"
	kindOptionComponents  = "components"
	kindOptionActions     = "actions"
	kindOptionAppBranches = "app_branches"

	subscribePredicateBlockID  = "nuon_subscribe_predicate_block"
	subscribePredicateActionID = "nuon_subscribe_predicate"
	predicateOptionAny         = "any"
	predicateOptionSpecific    = "specific"
	predicateOptionLabels      = "labels"

	subscribeAppBlockID  = "nuon_subscribe_app_block"
	subscribeAppActionID = "nuon_subscribe_app"

	subscribeEntitiesBlockID             = "nuon_subscribe_entities_block"
	subscribeEntitiesActionIDInstalls    = "nuon_subscribe_entities_installs"
	subscribeEntitiesActionIDComponents  = "nuon_subscribe_entities_components"
	subscribeEntitiesActionIDActions     = "nuon_subscribe_entities_actions"
	subscribeEntitiesActionIDAppBranches = "nuon_subscribe_entities_app_branches"

	subscribeLabelsBlockID         = "nuon_subscribe_labels_block"
	subscribeLabelsActionID        = "nuon_subscribe_labels"
	subscribeExcludeLabelsBlockID  = "nuon_subscribe_exclude_labels_block"
	subscribeExcludeLabelsActionID = "nuon_subscribe_exclude_labels"

	subscribeNotifBlockID  = "nuon_subscribe_notif_block"
	subscribeNotifActionID = "nuon_subscribe_notif"

	subscribeResourceBlockIDPrefix = "nuon_subscribe_res_"

	notifOptionAll      = "all"
	notifOptionSpecific = "specific"

	resourceOptEnable = "enable"

	categoryOptionLifecycle       = "lifecycle"
	categoryOptionApprovals       = "approvals"
	categoryOptionDrift           = "drift"
	categoryOptionComponentHealth = "component_health"
	categoryOptionInstallDegraded = "install_degraded"

	outcomeOptionNone       = "none"
	outcomeOptionAll        = "all"
	outcomeOptionCompletion = "completion"
	outcomeOptionFailures   = "failures"
)

func subscribeResourceOptsBlockID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_opts_block"
}

func subscribeResourceOptsActionID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_opts"
}

func subscribeResourceCategoriesBlockID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_categories_block"
}

func subscribeResourceCategoriesActionID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_categories"
}

func subscribeResourceLifecycleBlockID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_lifecycle_block"
}

func subscribeResourceLifecycleActionID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_lifecycle"
}

func subscribeResourceSubOpsBlockID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_subops_block"
}

func subscribeResourceSubOpsActionID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_subops"
}

type subscribeModalState struct {
	TeamID      string `json:"team_id"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	SlackUserID string `json:"slack_user_id"`
}

func encodeSubscribeModalState(s subscribeModalState) (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("encode subscribe modal state: %w", err)
	}
	return string(b), nil
}

func decodeSubscribeModalState(raw string) (subscribeModalState, error) {
	var s subscribeModalState
	if raw == "" {
		return s, errors.New("subscribe modal: empty private_metadata")
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return s, fmt.Errorf("decode subscribe modal state: %w", err)
	}
	return s, nil
}

type subscribeResourceCfg struct {
	Enabled   bool
	Ops       []string
	Approvals bool
	Drift     bool
	Outcome   string

	ComponentHealth bool
	InstallDegraded bool
}

type subscribeModalRenderState struct {
	OrgLinkID string

	Match string

	TargetKind labels.TargetKind

	AppID   string
	AppName string

	Predicate string

	EntityIDs   []string
	EntityNames []string

	LabelsRaw string

	ExcludeLabelsRaw string

	Notif string

	Resources map[interests.ResourceKind]subscribeResourceCfg
}

type subscribePreview struct {
	Kind  labels.TargetKind
	Count int
	Names []string
	Err   error
}

func buildSubscribeModalView(
	state subscribeModalState,
	links []app.SlackOrgLink,
	render subscribeModalRenderState,
	preview *subscribePreview,
) (map[string]any, error) {
	if len(links) == 0 {
		return nil, errors.New("subscribe modal: no verified org links")
	}

	if render.OrgLinkID == "" {
		render.OrgLinkID = links[0].ID
	}
	if render.Match == "" {
		render.Match = matchOptionAll
	}
	if render.Match == matchOptionSpecific {
		if render.TargetKind == "" {
			render.TargetKind = labels.TargetKindInstalls
		}
		if render.Predicate == "" {
			render.Predicate = predicateOptionAny
		}
	}
	if render.Notif == "" {
		render.Notif = notifOptionAll
	}

	pm, err := encodeSubscribeModalState(state)
	if err != nil {
		return nil, err
	}

	blocks := make([]any, 0, 8)

	channelLabel := state.ChannelID
	if state.ChannelName != "" {
		channelLabel = "#" + state.ChannelName
	}
	blocks = append(blocks, map[string]any{
		"type": "section",
		"text": map[string]any{
			"type": "mrkdwn",
			"text": fmt.Sprintf("Subscribe <#%s> (%s) to Nuon notifications.", state.ChannelID, channelLabel),
		},
	})

	orgOptions := make([]any, 0, len(links))
	var orgInitial map[string]any
	for _, link := range links {
		name := link.Org.Name
		if name == "" {
			name = link.OrgID
		}
		opt := map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": truncatePlainText(name, 75)},
			"value": link.ID,
		}
		orgOptions = append(orgOptions, opt)
		if link.ID == render.OrgLinkID {
			orgInitial = opt
		}
	}
	orgSelect := map[string]any{
		"type":      "static_select",
		"action_id": subscribeOrgActionID,
		"options":   orgOptions,
	}
	if orgInitial != nil {
		orgSelect["initial_option"] = orgInitial
	}
	blocks = append(blocks, map[string]any{
		"type":            "input",
		"block_id":        subscribeOrgBlockID,
		"label":           map[string]any{"type": "plain_text", "text": "Nuon org"},
		"element":         orgSelect,
		"dispatch_action": false,
	})

	matchOptAll := map[string]any{
		"text":  map[string]any{"type": "plain_text", "text": "Everything in this org"},
		"value": matchOptionAll,
	}
	matchOptSpecific := map[string]any{
		"text":  map[string]any{"type": "plain_text", "text": "Specific resources"},
		"value": matchOptionSpecific,
	}
	matchOptions := []any{matchOptAll, matchOptSpecific}
	matchInitial := matchOptAll
	if render.Match == matchOptionSpecific {
		matchInitial = matchOptSpecific
	}
	blocks = append(blocks, map[string]any{
		"type":     "input",
		"block_id": subscribeMatchBlockID,
		"label":    map[string]any{"type": "plain_text", "text": "What should this match?"},
		"element": map[string]any{
			"type":           "radio_buttons",
			"action_id":      subscribeMatchActionID,
			"initial_option": matchInitial,
			"options":        matchOptions,
		},
		"dispatch_action": true,
	})

	if render.Match == matchOptionSpecific {
		kindOptInstalls := map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "Installs"},
			"value": kindOptionInstalls,
		}
		kindOptComponents := map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "Components"},
			"value": kindOptionComponents,
		}
		kindOptActions := map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "Actions"},
			"value": kindOptionActions,
		}
		kindOptAppBranches := map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "App branches"},
			"value": kindOptionAppBranches,
		}
		kindOptions := []any{kindOptInstalls, kindOptComponents, kindOptActions, kindOptAppBranches}
		kindInitial := kindOptInstalls
		switch render.TargetKind {
		case labels.TargetKindComponents:
			kindInitial = kindOptComponents
		case labels.TargetKindActions:
			kindInitial = kindOptActions
		case labels.TargetKindAppBranches:
			kindInitial = kindOptAppBranches
		}
		blocks = append(blocks, map[string]any{
			"type":     "input",
			"block_id": subscribeKindBlockID,
			"label":    map[string]any{"type": "plain_text", "text": "Resource type"},
			"element": map[string]any{
				"type":           "static_select",
				"action_id":      subscribeKindActionID,
				"initial_option": kindInitial,
				"options":        kindOptions,
			},
			"dispatch_action": true,
		})

		predicateOptAny := map[string]any{
			"text":        map[string]any{"type": "plain_text", "text": "Any " + targetKindLabel(render.TargetKind, true)},
			"value":       predicateOptionAny,
			"description": map[string]any{"type": "plain_text", "text": "Match every " + targetKindLabel(render.TargetKind, false) + " in this org."},
		}
		predicateOptSpecific := map[string]any{
			"text":        map[string]any{"type": "plain_text", "text": "Specific " + targetKindLabel(render.TargetKind, true)},
			"value":       predicateOptionSpecific,
			"description": map[string]any{"type": "plain_text", "text": "Pick individual " + targetKindLabel(render.TargetKind, true) + " by name."},
		}
		predicateOptLabels := map[string]any{
			"text":        map[string]any{"type": "plain_text", "text": "By labels"},
			"value":       predicateOptionLabels,
			"description": map[string]any{"type": "plain_text", "text": "Match by label selector (env=prod, owner=*)."},
		}
		predicateOptions := []any{predicateOptAny, predicateOptSpecific, predicateOptLabels}
		predicateInitial := predicateOptAny
		switch render.Predicate {
		case predicateOptionSpecific:
			predicateInitial = predicateOptSpecific
		case predicateOptionLabels:
			predicateInitial = predicateOptLabels
		}
		blocks = append(blocks, map[string]any{
			"type":     "input",
			"block_id": subscribePredicateBlockID,
			"label":    map[string]any{"type": "plain_text", "text": "Match by"},
			"element": map[string]any{
				"type":           "radio_buttons",
				"action_id":      subscribePredicateActionID,
				"initial_option": predicateInitial,
				"options":        predicateOptions,
			},
			"dispatch_action": true,
		})

		switch render.Predicate {
		case predicateOptionSpecific:
			needsApp := targetKindNeedsApp(render.TargetKind)
			if needsApp {
				appSelect := map[string]any{
					"type":             "external_select",
					"action_id":        subscribeAppActionID,
					"min_query_length": 0,
					"placeholder":      map[string]any{"type": "plain_text", "text": "Search apps"},
				}
				if render.AppID != "" {
					appSelect["initial_option"] = map[string]any{
						"text": map[string]any{
							"type": "plain_text",
							"text": truncatePlainText(entityPickerLabel(render.AppID, render.AppName), 75),
						},
						"value": render.AppID,
					}
				}
				blocks = append(blocks, map[string]any{
					"type":            "input",
					"block_id":        subscribeAppBlockID,
					"label":           map[string]any{"type": "plain_text", "text": "App"},
					"element":         appSelect,
					"dispatch_action": true,
				})
			}

			if !needsApp || render.AppID != "" {
				actionID := subscribeEntitiesActionIDForKind(render.TargetKind)
				entSelect := map[string]any{
					"type":             "multi_external_select",
					"action_id":        actionID,
					"min_query_length": 0,
					"placeholder":      map[string]any{"type": "plain_text", "text": "Search " + targetKindLabel(render.TargetKind, true)},
				}
				if len(render.EntityIDs) > 0 {
					initial := make([]any, 0, len(render.EntityIDs))
					for i, id := range render.EntityIDs {
						name := ""
						if i < len(render.EntityNames) {
							name = render.EntityNames[i]
						}
						initial = append(initial, map[string]any{
							"text": map[string]any{
								"type": "plain_text",
								"text": truncatePlainText(entityPickerLabel(id, name), 75),
							},
							"value": id,
						})
					}
					entSelect["initial_options"] = initial
				}
				blocks = append(blocks, map[string]any{
					"type":            "input",
					"block_id":        subscribeEntitiesBlockID,
					"label":           map[string]any{"type": "plain_text", "text": titleCase(targetKindLabel(render.TargetKind, true))},
					"element":         entSelect,
					"dispatch_action": true,
				})
			}
		case predicateOptionLabels:
			labelsInput := map[string]any{
				"type":      "plain_text_input",
				"action_id": subscribeLabelsActionID,
				"placeholder": map[string]any{
					"type": "plain_text",
					"text": "env=prod, tier=critical, owner=*",
				},
			}
			if render.LabelsRaw != "" {
				labelsInput["initial_value"] = render.LabelsRaw
			}
			blocks = append(blocks, map[string]any{
				"type":            "input",
				"block_id":        subscribeLabelsBlockID,
				"label":           map[string]any{"type": "plain_text", "text": "Include labels"},
				"hint":            map[string]any{"type": "plain_text", "text": "Leave empty to include everything (combine with Exclude labels below)."},
				"optional":        true,
				"element":         labelsInput,
				"dispatch_action": true,
			})

			excludeInput := map[string]any{
				"type":      "plain_text_input",
				"action_id": subscribeExcludeLabelsActionID,
				"placeholder": map[string]any{
					"type": "plain_text",
					"text": "env=stage, tier=experimental",
				},
			}
			if render.ExcludeLabelsRaw != "" {
				excludeInput["initial_value"] = render.ExcludeLabelsRaw
			}
			blocks = append(blocks, map[string]any{
				"type":            "input",
				"block_id":        subscribeExcludeLabelsBlockID,
				"label":           map[string]any{"type": "plain_text", "text": "Exclude labels"},
				"hint":            map[string]any{"type": "plain_text", "text": "Skip events for entities matching these labels (e.g. env=stage)."},
				"optional":        true,
				"element":         excludeInput,
				"dispatch_action": true,
			})
		}

		if render.Predicate != predicateOptionAny && preview != nil {
			blocks = append(blocks, buildPreviewContextBlock(render.TargetKind, preview))
		}
	}

	notifOpts := []any{
		map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "All events"},
			"value": notifOptionAll,
		},
		map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "Specific events"},
			"value": notifOptionSpecific,
		},
	}
	notifInitial := notifOpts[0]
	if render.Notif == notifOptionSpecific {
		notifInitial = notifOpts[1]
	}
	blocks = append(blocks, map[string]any{
		"type":     "input",
		"block_id": subscribeNotifBlockID,
		"label":    map[string]any{"type": "plain_text", "text": "Notifications"},
		"element": map[string]any{
			"type":           "radio_buttons",
			"action_id":      subscribeNotifActionID,
			"initial_option": notifInitial,
			"options":        notifOpts,
		},
		"dispatch_action": true,
	})

	if render.Notif == notifOptionSpecific {
		seeded := seedResourcesForRender(render.Resources)
		blocks = append(blocks, map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": "*Per-resource filters*\nEnable the resources you care about, then pick which event categories (lifecycle, approvals, drift detection, health) you want notifications for.",
			},
		})
		for _, kind := range interests.AllResources {
			cfg := seeded[kind]
			blocks = append(blocks, map[string]any{"type": "divider"})
			blocks = append(blocks, buildResourceOptsBlock(kind, cfg))
			if cfg.Enabled {
				blocks = append(blocks, buildResourceCategoriesBlock(kind, cfg))
				if cfg.Outcome != outcomeOptionNone {
					blocks = append(blocks, buildResourceLifecycleBlock(kind, cfg))
					blocks = append(blocks, buildResourceSubOpsBlock(kind, cfg))
				}
			}
		}
	}

	return map[string]any{
		"type":             "modal",
		"callback_id":      subscribeModalCallbackID,
		"private_metadata": pm,
		"title":            map[string]any{"type": "plain_text", "text": "Subscribe to Nuon"},
		"submit":           map[string]any{"type": "plain_text", "text": "Subscribe"},
		"close":            map[string]any{"type": "plain_text", "text": "Cancel"},
		"blocks":           blocks,
	}, nil
}

func resourceKindLabel(k interests.ResourceKind) string {
	switch k {
	case interests.ResourceInstalls:
		return "Installs"
	case interests.ResourceStacks:
		return "Stacks"
	case interests.ResourceComponents:
		return "Components"
	case interests.ResourceSandboxes:
		return "Sandboxes"
	case interests.ResourceInstallConfigurations:
		return "Install configurations"
	case interests.ResourceRunners:
		return "Runners"
	case interests.ResourceActions:
		return "Actions"
	default:
		return string(k)
	}
}

func subOpLabel(op string) string {
	switch op {
	case "provision":
		return "Provision"
	case "deprovision":
		return "Deprovision"
	case "reprovision":
		return "Reprovision"
	case "deploy":
		return "Deploy"
	case "teardown":
		return "Teardown"
	case "inputs":
		return "Inputs"
	case "secrets":
		return "Secrets"
	case "inactive":
		return "Inactive"
	case "unhealthy":
		return "Unhealthy"
	case "run":
		return "Run"
	case "version_active":
		return "Version active"
	default:
		return op
	}
}

func seedResourcesForRender(in map[interests.ResourceKind]subscribeResourceCfg) map[interests.ResourceKind]subscribeResourceCfg {
	if len(in) > 0 {
		return in
	}
	defaults := interests.Default().Resources
	out := make(map[interests.ResourceKind]subscribeResourceCfg, len(interests.AllResources))
	for _, kind := range interests.AllResources {
		cfg, ok := defaults[kind]
		if !ok {
			out[kind] = subscribeResourceCfg{Outcome: outcomeOptionCompletion}
			continue
		}
		out[kind] = subscribeResourceCfg{
			Enabled:         true,
			Ops:             append([]string(nil), cfg.Ops...),
			Approvals:       cfg.ApprovalRequests || cfg.ApprovalResponses,
			Drift:           cfg.DriftDetected,
			ComponentHealth: cfg.ComponentHealth,
			InstallDegraded: cfg.InstallDegraded,
			Outcome:         outcomeFromInterests(cfg.Outcome),
		}
	}
	return out
}

func outcomeFromInterests(o interests.Outcome) string {
	switch o {
	case interests.OutcomeNone:
		return outcomeOptionNone
	case interests.OutcomeCompletion:
		return outcomeOptionCompletion
	case interests.OutcomeFailures:
		return outcomeOptionFailures
	default:
		return outcomeOptionAll
	}
}

// why: outcomeToInterests is the inverse of outcomeFromInterests. Unknown values
// (including "" which Slack will only emit if the radio truly has no
// initial_option and the user never touched it) fall back to OutcomeAll so
// we never persist an empty Outcome that means something different to the
// matcher than the user expected.
func outcomeToInterests(s string) interests.Outcome {
	switch s {
	case outcomeOptionNone:
		return interests.OutcomeNone
	case outcomeOptionCompletion:
		return interests.OutcomeCompletion
	case outcomeOptionFailures:
		return interests.OutcomeFailures
	default:
		return interests.OutcomeAll
	}
}

func outcomeRadioOptions() []any {
	return []any{
		map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "All events"},
			"value": outcomeOptionAll,
		},
		map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "On completion"},
			"value": outcomeOptionCompletion,
		},
		map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "On failures"},
			"value": outcomeOptionFailures,
		},
	}
}

func buildResourceOptsBlock(kind interests.ResourceKind, cfg subscribeResourceCfg) map[string]any {
	enableOpt := map[string]any{
		"text":  map[string]any{"type": "plain_text", "text": "Enable " + resourceKindLabel(kind)},
		"value": resourceOptEnable,
	}
	element := map[string]any{
		"type":      "checkboxes",
		"action_id": subscribeResourceOptsActionID(kind),
		"options":   []any{enableOpt},
	}
	if cfg.Enabled {
		element["initial_options"] = []any{enableOpt}
	}
	return map[string]any{
		"type":            "input",
		"block_id":        subscribeResourceOptsBlockID(kind),
		"label":           map[string]any{"type": "plain_text", "text": resourceKindLabel(kind)},
		"element":         element,
		"optional":        true,
		"dispatch_action": true,
	}
}

func buildResourceLifecycleBlock(kind interests.ResourceKind, cfg subscribeResourceCfg) map[string]any {
	opts := outcomeRadioOptions()
	want := cfg.Outcome
	if want == "" || want == outcomeOptionNone {
		want = outcomeOptionCompletion
	}
	initial := opts[0]
	for _, o := range opts {
		m := o.(map[string]any)
		if m["value"] == want {
			initial = m
			break
		}
	}
	return map[string]any{
		"type":     "input",
		"block_id": subscribeResourceLifecycleBlockID(kind),
		"label":    map[string]any{"type": "plain_text", "text": resourceKindLabel(kind) + " — lifecycle events"},
		"optional": true,
		"element": map[string]any{
			"type":           "radio_buttons",
			"action_id":      subscribeResourceLifecycleActionID(kind),
			"options":        opts,
			"initial_option": initial,
		},
	}
}

func buildResourceCategoriesBlock(kind interests.ResourceKind, cfg subscribeResourceCfg) map[string]any {
	lifecycleOpt := map[string]any{
		"text":  map[string]any{"type": "plain_text", "text": "Lifecycle events"},
		"value": categoryOptionLifecycle,
	}
	approvalsOpt := map[string]any{
		"text":  map[string]any{"type": "plain_text", "text": "Approval events"},
		"value": categoryOptionApprovals,
	}
	options := []any{lifecycleOpt, approvalsOpt}

	var initial []any
	if cfg.Outcome != outcomeOptionNone {
		initial = append(initial, lifecycleOpt)
	}
	if cfg.Approvals {
		initial = append(initial, approvalsOpt)
	}
	if interests.SupportsDriftDetected(kind) {
		driftOpt := map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "Drift detection"},
			"value": categoryOptionDrift,
		}
		options = append(options, driftOpt)
		if cfg.Drift {
			initial = append(initial, driftOpt)
		}
	}
	if interests.SupportsComponentHealth(kind) {
		healthOpt := map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "Component health"},
			"value": categoryOptionComponentHealth,
		}
		options = append(options, healthOpt)
		if cfg.ComponentHealth {
			initial = append(initial, healthOpt)
		}
	}
	if interests.SupportsInstallDegraded(kind) {
		degradedOpt := map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": "Install degraded"},
			"value": categoryOptionInstallDegraded,
		}
		options = append(options, degradedOpt)
		if cfg.InstallDegraded {
			initial = append(initial, degradedOpt)
		}
	}

	element := map[string]any{
		"type":      "checkboxes",
		"action_id": subscribeResourceCategoriesActionID(kind),
		"options":   options,
	}
	if len(initial) > 0 {
		element["initial_options"] = initial
	}
	return map[string]any{
		"type":            "input",
		"block_id":        subscribeResourceCategoriesBlockID(kind),
		"label":           map[string]any{"type": "plain_text", "text": resourceKindLabel(kind) + " — event categories"},
		"element":         element,
		"optional":        true,
		"dispatch_action": true,
	}
}

func buildResourceSubOpsBlock(kind interests.ResourceKind, cfg subscribeResourceCfg) map[string]any {
	subOps := interests.SubOps[kind]
	options := make([]any, 0, len(subOps))
	opSet := make(map[string]struct{}, len(cfg.Ops))
	for _, op := range cfg.Ops {
		opSet[op] = struct{}{}
	}
	var initial []any
	for _, op := range subOps {
		opt := map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": subOpLabel(op)},
			"value": op,
		}
		options = append(options, opt)
		if _, ok := opSet[op]; ok {
			initial = append(initial, opt)
		}
	}
	element := map[string]any{
		"type":      "checkboxes",
		"action_id": subscribeResourceSubOpsActionID(kind),
		"options":   options,
	}
	if len(initial) > 0 {
		element["initial_options"] = initial
	}
	return map[string]any{
		"type":     "input",
		"block_id": subscribeResourceSubOpsBlockID(kind),
		"label":    map[string]any{"type": "plain_text", "text": resourceKindLabel(kind) + " — sub-operations"},
		"element":  element,
		"optional": true,
	}
}

func entityPickerLabel(id, name string) string {
	if name == "" {
		return id
	}
	return fmt.Sprintf("%s (%s)", name, id)
}

func targetKindLabel(k labels.TargetKind, plural bool) string {
	switch k {
	case labels.TargetKindComponents:
		if plural {
			return "components"
		}
		return "component"
	case labels.TargetKindActions:
		if plural {
			return "actions"
		}
		return "action"
	case labels.TargetKindAppBranches:
		if plural {
			return "app branches"
		}
		return "app branch"
	default:
		if plural {
			return "installs"
		}
		return "install"
	}
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func targetKindNeedsApp(k labels.TargetKind) bool {
	switch k {
	case labels.TargetKindComponents, labels.TargetKindActions, labels.TargetKindAppBranches:
		return true
	default:
		return false
	}
}

func subscribeEntitiesActionIDForKind(k labels.TargetKind) string {
	switch k {
	case labels.TargetKindComponents:
		return subscribeEntitiesActionIDComponents
	case labels.TargetKindActions:
		return subscribeEntitiesActionIDActions
	case labels.TargetKindAppBranches:
		return subscribeEntitiesActionIDAppBranches
	default:
		return subscribeEntitiesActionIDInstalls
	}
}

// why: targetKindFromString maps a kindOption* value back to a labels.TargetKind.
// Unknown values fall through to TargetKindInstalls so a tampered payload
// can't smuggle an unknown kind into the matcher.
func targetKindFromString(s string) labels.TargetKind {
	switch s {
	case kindOptionComponents:
		return labels.TargetKindComponents
	case kindOptionActions:
		return labels.TargetKindActions
	case kindOptionAppBranches:
		return labels.TargetKindAppBranches
	default:
		return labels.TargetKindInstalls
	}
}

func buildPreviewContextBlock(kind labels.TargetKind, p *subscribePreview) map[string]any {
	var text string
	switch {
	case p.Err != nil:
		text = "_Couldn't compute preview: " + p.Err.Error() + "_"
	case p.Count == 0:
		text = ":warning: No " + targetKindLabel(kind, true) + " currently match this filter."
	default:
		const maxNames = 6
		shown := p.Names
		if len(shown) > maxNames {
			shown = shown[:maxNames]
		}
		joined := strings.Join(shown, ", ")
		text = fmt.Sprintf("Currently matches *%d* %s: %s", p.Count, targetKindLabel(kind, true), joined)
		if p.Count > len(shown) {
			text += fmt.Sprintf(", +%d more", p.Count-len(shown))
		}
	}
	return map[string]any{
		"type": "context",
		"elements": []any{
			map[string]any{
				"type": "mrkdwn",
				"text": text,
			},
		},
	}
}

func truncatePlainText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	const ellipsis = "…"
	if max <= len(ellipsis) {
		return safeBytePrefix(s, max)
	}
	return safeBytePrefix(s, max-len(ellipsis)) + ellipsis
}

// why: safeBytePrefix returns the longest prefix of s whose UTF-8 length is
// <= n bytes, never splitting a multi-byte rune. Slack rejects invalid
// UTF-8 in plain_text so a naïve s[:n] on a multi-byte string would be
// unsafe.
func safeBytePrefix(s string, n int) string {
	if n <= 0 || n >= len(s) {
		if n >= len(s) {
			return s
		}
		return ""
	}
	out := 0
	for i := range s {
		if i > n {
			break
		}
		out = i
	}
	return s[:out]
}

func (s *service) openSubscribeModalForSlash(
	ctx context.Context,
	triggerID string,
	teamID, channelID, channelName, slackUserID string,
	preselect subscribeModalRenderState,
) error {
	install, err := s.lookupActiveInstallForTeam(ctx, teamID)
	if err != nil {
		return err
	}

	var links []app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Preload("Org").
		Where(app.SlackOrgLink{TeamID: teamID, Status: app.SlackOrgLinkStatusVerified}).
		Find(&links).Error; err != nil {
		return fmt.Errorf("subscribe modal: list org links: %w", err)
	}
	if len(links) == 0 {
		return errors.New("subscribe modal: no verified org links for workspace")
	}

	if preselect.OrgLinkID != "" {
		known := false
		for _, l := range links {
			if l.ID == preselect.OrgLinkID {
				known = true
				break
			}
		}
		if !known {
			preselect = subscribeModalRenderState{}
		}
	}

	state := subscribeModalState{
		TeamID:      teamID,
		ChannelID:   channelID,
		ChannelName: channelName,
		SlackUserID: slackUserID,
	}

	var orgID string
	for _, l := range links {
		if l.ID == preselect.OrgLinkID {
			orgID = l.OrgID
			break
		}
	}
	preview := s.maybePreview(ctx, orgID, preselect)

	view, err := buildSubscribeModalView(state, links, preselect, preview)
	if err != nil {
		return err
	}

	if _, err := s.slackClient.ViewsOpen(ctx, install.BotAccessToken, slackclient.ViewsOpenRequest{
		TriggerID: triggerID,
		View:      view,
	}); err != nil {
		return fmt.Errorf("subscribe modal: views.open: %w", err)
	}
	return nil
}

func (s *service) preselectSubscribeRenderStateForChannel(
	ctx context.Context,
	teamID, channelID string,
) subscribeModalRenderState {
	var sub app.SlackChannelSubscription
	if err := s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{TeamID: teamID, ChannelID: channelID}).
		Order("updated_at DESC").
		First(&sub).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.l.Warn("subscribe modal: lookup existing subscription for preselect failed",
				zap.Error(err),
				zap.String("team_id", teamID),
				zap.String("channel_id", channelID))
		}
		return subscribeModalRenderState{}
	}
	return s.renderStateFromSubscription(ctx, sub)
}

func (s *service) renderStateFromSubscription(
	ctx context.Context,
	sub app.SlackChannelSubscription,
) subscribeModalRenderState {
	rs := subscribeModalRenderState{
		OrgLinkID: sub.OrgLinkID,
	}

	if sub.Match == nil || (sub.Match.Installs == nil && sub.Match.Components == nil && sub.Match.Actions == nil) {
		rs.Match = matchOptionAll
	} else {
		rs.Match = matchOptionSpecific
		var (
			kind labels.TargetKind
			tm   *labels.TargetMatch
		)
		switch {
		case sub.Match.Installs != nil:
			kind, tm = labels.TargetKindInstalls, sub.Match.Installs
		case sub.Match.Components != nil:
			kind, tm = labels.TargetKindComponents, sub.Match.Components
		case sub.Match.Actions != nil:
			kind, tm = labels.TargetKindActions, sub.Match.Actions
		case sub.Match.AppBranches != nil:
			kind, tm = labels.TargetKindAppBranches, sub.Match.AppBranches
		}
		rs.TargetKind = kind
		switch {
		case len(tm.IDs) > 0:
			rs.Predicate = predicateOptionSpecific
			rs.EntityIDs = append([]string(nil), tm.IDs...)
			rs.EntityNames = s.lookupEntityNames(ctx, sub.OrgID, kind, tm.IDs)
			if targetKindNeedsApp(kind) {
				rs.AppID, rs.AppName = s.lookupOwningApp(ctx, sub.OrgID, kind, tm.IDs[0])
			}
		case tm.Selector != nil:
			rs.Predicate = predicateOptionLabels
			rs.LabelsRaw = labelsToQueryString(tm.Selector.MatchLabels)
			rs.ExcludeLabelsRaw = labelsToQueryString(tm.Selector.NotMatchLabels)
		default:
			rs.Predicate = predicateOptionAny
		}
	}

	if sub.Interests.AllEvents {
		rs.Notif = notifOptionAll
	} else if sub.Interests.IsZero() {
		rs.Notif = notifOptionAll
	} else {
		rs.Notif = notifOptionSpecific
		rs.Resources = renderResourcesFromInterests(sub.Interests)
	}
	return rs
}

func describeMatch(m *labels.SubscriptionMatch) string {
	if m == nil {
		return "everything in org"
	}
	var (
		kindLabel string
		tm        *labels.TargetMatch
	)
	switch {
	case m.Installs != nil:
		kindLabel, tm = "installs", m.Installs
	case m.Components != nil:
		kindLabel, tm = "components", m.Components
	case m.Actions != nil:
		kindLabel, tm = "actions", m.Actions
	default:
		return "everything in org"
	}
	switch {
	case tm.Selector != nil:
		parts := make([]string, 0, 2)
		if inc := labelsToQueryString(tm.Selector.MatchLabels); inc != "" {
			parts = append(parts, inc)
		}
		if exc := labelsToQueryString(tm.Selector.NotMatchLabels); exc != "" {
			parts = append(parts, "not "+exc)
		}
		return "by labels: " + strings.Join(parts, "; ")
	case len(tm.IDs) > 0:
		const maxIDs = 3
		shown := tm.IDs
		if len(shown) > maxIDs {
			shown = shown[:maxIDs]
		}
		out := "specific " + kindLabel + ": " + strings.Join(shown, ", ")
		if len(tm.IDs) > len(shown) {
			out += fmt.Sprintf(", +%d more", len(tm.IDs)-len(shown))
		}
		return out
	default:
		return "any " + kindLabel
	}
}

func labelsToQueryString(m labels.Labels) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := m[k]
		if v == "*" {
			parts = append(parts, k+"=*")
		} else {
			parts = append(parts, k+"="+v)
		}
	}
	return strings.Join(parts, ", ")
}

func (s *service) lookupEntityNames(
	ctx context.Context,
	orgID string,
	kind labels.TargetKind,
	ids []string,
) []string {
	if s.db == nil || orgID == "" || len(ids) == 0 {
		return nil
	}
	table := tableForTargetKind(kind)
	if table == "" {
		return nil
	}
	type row struct {
		ID   string
		Name string
	}
	var rows []row
	if err := s.db.WithContext(ctx).
		Table(table).
		Select("id, name").
		Where("org_id = ? AND id IN ?", orgID, ids).
		Find(&rows).Error; err != nil {
		s.l.Warn("subscribe modal: lookup entity names failed",
			zap.Error(err),
			zap.String("kind", string(kind)))
		return make([]string, len(ids))
	}
	byID := make(map[string]string, len(rows))
	for _, r := range rows {
		byID[r.ID] = r.Name
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = byID[id]
	}
	return out
}

func (s *service) lookupOwningApp(
	ctx context.Context,
	orgID string,
	kind labels.TargetKind,
	entityID string,
) (string, string) {
	if s.db == nil || orgID == "" || entityID == "" {
		return "", ""
	}
	table := tableForTargetKind(kind)
	if table == "" || !targetKindNeedsApp(kind) {
		return "", ""
	}
	type row struct {
		AppID   string `gorm:"column:app_id"`
		AppName string `gorm:"column:app_name"`
	}
	var r row
	if err := s.db.WithContext(ctx).
		Table(table+" AS e").
		Select("e.app_id AS app_id, a.name AS app_name").
		Joins("JOIN apps AS a ON a.id = e.app_id").
		Where("e.org_id = ? AND e.id = ?", orgID, entityID).
		Take(&r).Error; err != nil {
		s.l.Warn("subscribe modal: lookup owning app failed",
			zap.Error(err),
			zap.String("kind", string(kind)),
			zap.String("entity_id", entityID))
		return "", ""
	}
	return r.AppID, r.AppName
}

func tableForTargetKind(kind labels.TargetKind) string {
	switch kind {
	case labels.TargetKindInstalls:
		return "installs"
	case labels.TargetKindComponents:
		return "components"
	case labels.TargetKindActions:
		return "action_workflows"
	case labels.TargetKindAppBranches:
		return "app_branches"
	default:
		return ""
	}
}

func renderResourcesFromInterests(in interests.Interests) map[interests.ResourceKind]subscribeResourceCfg {
	out := make(map[interests.ResourceKind]subscribeResourceCfg, len(interests.AllResources))
	for _, kind := range interests.AllResources {
		if cfg, ok := in.Resources[kind]; ok {
			out[kind] = subscribeResourceCfg{
				Enabled:         true,
				Ops:             append([]string(nil), cfg.Ops...),
				Approvals:       cfg.ApprovalRequests || cfg.ApprovalResponses,
				Drift:           cfg.DriftDetected,
				ComponentHealth: cfg.ComponentHealth,
				InstallDegraded: cfg.InstallDegraded,
				Outcome:         outcomeFromInterests(cfg.Outcome),
			}
			continue
		}
		out[kind] = subscribeResourceCfg{Outcome: outcomeOptionCompletion}
	}
	return out
}

func (s *service) lookupActiveInstallForTeam(ctx context.Context, teamID string) (*app.SlackInstallation, error) {
	var install app.SlackInstallation
	if err := s.db.WithContext(ctx).
		Where(app.SlackInstallation{TeamID: teamID, Status: app.SlackInstallationStatusActive}).
		First(&install).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subscribe modal: workspace %q has no active install", teamID)
		}
		return nil, fmt.Errorf("subscribe modal: lookup installation: %w", err)
	}
	return &install, nil
}

func (s *service) rerenderSubscribeModal(
	ctx context.Context,
	payload slackInteractionPayload,
	render subscribeModalRenderState,
) error {
	state, err := decodeSubscribeModalState(payload.View.PrivateMetadata)
	if err != nil {
		return err
	}
	if state.TeamID != payload.Team.ID {
		return fmt.Errorf("subscribe modal: team_id mismatch (state=%q, payload=%q)", state.TeamID, payload.Team.ID)
	}

	install, err := s.lookupActiveInstallForTeam(ctx, state.TeamID)
	if err != nil {
		return err
	}
	var links []app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Preload("Org").
		Where(app.SlackOrgLink{TeamID: state.TeamID, Status: app.SlackOrgLinkStatusVerified}).
		Find(&links).Error; err != nil {
		return fmt.Errorf("subscribe modal: list org links: %w", err)
	}

	var orgID string
	for _, l := range links {
		if l.ID == render.OrgLinkID {
			orgID = l.OrgID
			break
		}
	}
	preview := s.maybePreview(ctx, orgID, render)

	view, err := buildSubscribeModalView(state, links, render, preview)
	if err != nil {
		return err
	}
	if _, err := s.slackClient.ViewsUpdate(ctx, install.BotAccessToken, slackclient.ViewsUpdateRequest{
		ViewID: payload.View.ID,
		Hash:   payload.View.Hash,
		View:   view,
	}); err != nil {
		return fmt.Errorf("subscribe modal: views.update: %w", err)
	}
	return nil
}

func (s *service) handleSubscribeModalBlockActions(ctx context.Context, payload slackInteractionPayload) {
	render := readSubscribeRenderStateFromPayload(payload)

	// why: Switching the app means the prior entity selections refer to
	// resources outside the new app's scope. The entity picker's
	// block_id is stable across app changes, so Slack would otherwise
	// echo the stale ids back in view.state.values. Drop them here.
	for _, a := range payload.Actions {
		if a.ActionID == subscribeAppActionID {
			render.EntityIDs = nil
			render.EntityNames = nil
			break
		}
	}

	if err := s.rerenderSubscribeModal(ctx, payload, render); err != nil {
		s.l.Warn("subscribe modal: re-render failed", zap.Error(err))
	}
}

func (s *service) handleSubscribeModalSubmission(ctx context.Context, payload slackInteractionPayload) any {
	state, err := decodeSubscribeModalState(payload.View.PrivateMetadata)
	if err != nil {
		s.l.Warn("subscribe modal submit: decode state failed", zap.Error(err))
		return modalErr(subscribeOrgBlockID, "Internal error — please re-open the modal.")
	}
	if state.TeamID != payload.Team.ID {
		s.l.Warn("subscribe modal submit: team_id mismatch",
			zap.String("state_team_id", state.TeamID),
			zap.String("payload_team_id", payload.Team.ID))
		return modalErr(subscribeOrgBlockID, "Internal error — please re-open the modal.")
	}

	render := readSubscribeRenderStateFromPayload(payload)

	var link app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{
			ID:     render.OrgLinkID,
			TeamID: state.TeamID,
			Status: app.SlackOrgLinkStatusVerified,
		}).
		First(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return modalErr(subscribeOrgBlockID, "That org is no longer linked to this workspace.")
		}
		s.l.Error("subscribe modal submit: lookup org link failed", zap.Error(err))
		return modalErr(subscribeOrgBlockID, "Sorry — something went wrong looking up the org link.")
	}

	match, errBlock, errMsg := buildSubscriptionMatchFromRender(render)
	if errMsg != "" {
		return modalErr(errBlock, errMsg)
	}

	// why: Anti-tampering: when Predicate=Specific, all chosen entity IDs
	// must belong to the linked org. The block_suggestion handler
	// already filters options to link.OrgID, but a determined attacker
	// could craft a view_submission with foreign ids. For component /
	// action kinds the picker is further scoped to a chosen app — the
	// validator re-derives that boundary so a tampered payload can't
	// smuggle in entities from a sibling app.
	if match != nil {
		if errBlock, errMsg := s.validateMatchEntityIDs(ctx, link.OrgID, render.AppID, match); errMsg != "" {
			return modalErr(errBlock, errMsg)
		}
	}

	var in interests.Interests
	if render.Notif == notifOptionSpecific {
		in = buildSpecificEventsInterests(render.Resources)
	} else {
		in = interests.Interests{AllEvents: true}
	}

	if err := s.upsertModalSubscription(ctx, link, state, match, in); err != nil {
		s.l.Error("subscribe modal submit: upsert subscription failed", zap.Error(err))
		return modalErr(subscribeOrgBlockID, "Sorry — couldn't save the subscription. Please try again.")
	}

	return map[string]any{}
}

func buildSubscriptionMatchFromRender(render subscribeModalRenderState) (*labels.SubscriptionMatch, string, string) {
	if render.Match != matchOptionSpecific {
		return nil, "", ""
	}

	kind := render.TargetKind
	if kind == "" {
		kind = labels.TargetKindInstalls
	}

	var tm *labels.TargetMatch
	switch render.Predicate {
	case predicateOptionSpecific:
		if targetKindNeedsApp(kind) && render.AppID == "" {
			return nil, subscribeAppBlockID, "Pick an app to choose " + targetKindLabel(kind, true) + " from."
		}
		if len(render.EntityIDs) == 0 {
			return nil, subscribeEntitiesBlockID, "Pick at least one " + targetKindLabel(kind, false) + " or change the match type."
		}
		tm = &labels.TargetMatch{IDs: append([]string(nil), render.EntityIDs...)}
	case predicateOptionLabels:
		includeLbls := labels.ParseLabelsQuery(render.LabelsRaw)
		excludeLbls := labels.ParseLabelsQuery(render.ExcludeLabelsRaw)
		if len(includeLbls) == 0 && len(excludeLbls) == 0 {
			return nil, subscribeLabelsBlockID, "Enter a label selector (include or exclude) like env=prod or env=stage."
		}
		sel := &labels.Selector{MatchLabels: includeLbls, NotMatchLabels: excludeLbls}
		if err := sel.Validate(); err != nil {
			return nil, subscribeLabelsBlockID, "Invalid label selector: " + err.Error()
		}
		tm = &labels.TargetMatch{Selector: sel}
	default:
		tm = &labels.TargetMatch{}
	}

	out := &labels.SubscriptionMatch{}
	switch kind {
	case labels.TargetKindComponents:
		out.Components = tm
	case labels.TargetKindActions:
		out.Actions = tm
	case labels.TargetKindAppBranches:
		out.AppBranches = tm
	default:
		out.Installs = tm
	}
	if err := out.Validate(); err != nil {
		return nil, subscribeEntitiesBlockID, "Invalid filter: " + err.Error()
	}
	return out, "", ""
}

func (s *service) validateMatchEntityIDs(ctx context.Context, orgID, appID string, m *labels.SubscriptionMatch) (string, string) {
	check := func(kind labels.TargetKind, tm *labels.TargetMatch) (string, string) {
		if tm == nil || len(tm.IDs) == 0 {
			return "", ""
		}
		table := tableForTargetKind(kind)
		if table == "" {
			return subscribeEntitiesBlockID, "Internal error: unknown resource kind."
		}
		tx := s.db.WithContext(ctx).
			Table(table).
			Where("org_id = ? AND id IN ?", orgID, tm.IDs)
		if targetKindNeedsApp(kind) && appID != "" {
			tx = tx.Where("app_id = ?", appID)
		}
		var count int64
		if err := tx.Count(&count).Error; err != nil {
			s.l.Error("subscribe modal submit: validate entity ids failed",
				zap.Error(err), zap.String("kind", string(kind)))
			return subscribeEntitiesBlockID, "Sorry — couldn't validate the chosen " + targetKindLabel(kind, true) + "."
		}
		if int(count) != len(tm.IDs) {
			return subscribeEntitiesBlockID, "One or more chosen " + targetKindLabel(kind, true) + " no longer exist in the chosen app."
		}
		return "", ""
	}
	if b, msg := check(labels.TargetKindInstalls, m.Installs); msg != "" {
		return b, msg
	}
	if b, msg := check(labels.TargetKindComponents, m.Components); msg != "" {
		return b, msg
	}
	if b, msg := check(labels.TargetKindActions, m.Actions); msg != "" {
		return b, msg
	}
	if b, msg := check(labels.TargetKindAppBranches, m.AppBranches); msg != "" {
		return b, msg
	}
	return "", ""
}

func buildSpecificEventsInterests(in map[interests.ResourceKind]subscribeResourceCfg) interests.Interests {
	out := interests.Interests{
		Resources: make(map[interests.ResourceKind]interests.ResourceCfg),
	}
	for _, kind := range interests.AllResources {
		cfg, ok := in[kind]
		if !ok || !cfg.Enabled {
			continue
		}

		// why: Filter sub-ops to the canonical vocabulary so a tampered
		// payload can't smuggle unknown slugs into the matcher.
		validOps := make(map[string]struct{}, len(interests.SubOps[kind]))
		for _, op := range interests.SubOps[kind] {
			validOps[op] = struct{}{}
		}
		ops := make([]string, 0, len(cfg.Ops))
		for _, op := range cfg.Ops {
			if _, ok := validOps[op]; ok {
				ops = append(ops, op)
			}
		}

		rc := interests.ResourceCfg{
			Outcome:           outcomeToInterests(cfg.Outcome),
			ApprovalRequests:  cfg.Approvals,
			ApprovalResponses: cfg.Approvals,
		}
		if len(ops) > 0 {
			rc.Ops = ops
		}
		if cfg.Drift && interests.SupportsDriftDetected(kind) {
			rc.DriftDetected = true
		}
		if cfg.ComponentHealth && interests.SupportsComponentHealth(kind) {
			rc.ComponentHealth = true
		}
		if cfg.InstallDegraded && interests.SupportsInstallDegraded(kind) {
			rc.InstallDegraded = true
		}
		out.Resources[kind] = rc
	}
	return out
}

func (s *service) upsertModalSubscription(
	ctx context.Context,
	link app.SlackOrgLink,
	state subscribeModalState,
	match *labels.SubscriptionMatch,
	in interests.Interests,
) error {
	slackUserID := state.SlackUserID
	sub := app.SlackChannelSubscription{
		OrgLinkID:            link.ID,
		OrgID:                link.OrgID,
		TeamID:               state.TeamID,
		ChannelID:            state.ChannelID,
		ChannelName:          state.ChannelName,
		Match:                match,
		Interests:            in,
		CreatedBySlackUserID: &slackUserID,
	}
	// why: The slack-side flow has no account context, so satisfy the
	// CreatedByID NOT NULL constraint with a deterministic placeholder.
	// SlackUserID is namespace-prefixed by Slack so it can never collide
	// with a real account id.
	sub.CreatedByID = "slack:" + slackUserID

	tx := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "team_id"},
			{Name: "channel_id"},
			{Name: "org_link_id"},
			{Name: "match_canonical"},
			{Name: "deleted_at"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"channel_name",
			"interests",
			"updated_at",
			"created_by_slack_user_id",
		}),
	}).Create(&sub)
	if err := tx.Error; err != nil {
		return pkgerrors.Wrap(err, "upsert slack channel subscription")
	}
	return nil
}

func (s *service) maybePreview(
	ctx context.Context,
	orgID string,
	render subscribeModalRenderState,
) *subscribePreview {
	if render.Match != matchOptionSpecific || render.Predicate == predicateOptionAny {
		return nil
	}
	if orgID == "" || s.db == nil {
		return nil
	}
	kind := render.TargetKind
	if kind == "" {
		kind = labels.TargetKindInstalls
	}

	var tm *labels.TargetMatch
	switch render.Predicate {
	case predicateOptionSpecific:
		if len(render.EntityIDs) == 0 {
			return &subscribePreview{Kind: kind}
		}
		tm = &labels.TargetMatch{IDs: append([]string(nil), render.EntityIDs...)}
	case predicateOptionLabels:
		includeLbls := labels.ParseLabelsQuery(render.LabelsRaw)
		excludeLbls := labels.ParseLabelsQuery(render.ExcludeLabelsRaw)
		if len(includeLbls) == 0 && len(excludeLbls) == 0 {
			return &subscribePreview{Kind: kind}
		}
		tm = &labels.TargetMatch{Selector: &labels.Selector{
			MatchLabels:    includeLbls,
			NotMatchLabels: excludeLbls,
		}}
	default:
		return nil
	}

	count, names, err := s.previewMatched(ctx, orgID, kind, tm)
	return &subscribePreview{Kind: kind, Count: count, Names: names, Err: err}
}

func (s *service) previewMatched(
	ctx context.Context,
	orgID string,
	kind labels.TargetKind,
	t *labels.TargetMatch,
) (int, []string, error) {
	if t == nil || (len(t.IDs) == 0 && t.Selector == nil) {
		return 0, nil, nil
	}
	table := tableForTargetKind(kind)
	if table == "" {
		return 0, nil, pkgerrors.Errorf("unknown target kind %q", kind)
	}

	const previewLimit = 7
	type row struct {
		Name string
	}

	tx := s.db.WithContext(ctx).
		Table(table).
		Where("org_id = ?", orgID)

	switch {
	case len(t.IDs) > 0:
		tx = tx.Where("id IN ?", t.IDs)
	case t.Selector != nil:
		tx = tx.Scopes(
			labels.WithLabels("labels", t.Selector.MatchLabels),
			labels.WithoutLabels("labels", t.Selector.NotMatchLabels),
		)
	}

	var total int64
	if err := tx.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return 0, nil, pkgerrors.Wrap(err, "preview count")
	}
	var rows []row
	if err := tx.Session(&gorm.Session{}).
		Select("name").
		Order("name ASC").
		Limit(previewLimit).
		Scan(&rows).Error; err != nil {
		return 0, nil, pkgerrors.Wrap(err, "preview names")
	}
	names := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.Name != "" {
			names = append(names, r.Name)
		}
	}
	return int(total), names, nil
}

func modalErr(blockID, message string) map[string]any {
	return map[string]any{
		"response_action": "errors",
		"errors": map[string]string{
			blockID: message,
		},
	}
}

func readSubscribeRenderStateFromPayload(payload slackInteractionPayload) subscribeModalRenderState {
	values := payload.View.State.Values
	get := func(blockID, actionID string) map[string]any {
		block, ok := values[blockID]
		if !ok {
			return nil
		}
		raw, ok := block[actionID]
		if !ok {
			return nil
		}
		return raw
	}
	pickValue := func(blockID, actionID, key string) string {
		raw := get(blockID, actionID)
		if raw == nil {
			return ""
		}
		opt, ok := raw[key].(map[string]any)
		if !ok {
			return ""
		}
		v, _ := opt["value"].(string)
		return v
	}
	pickMultiValues := func(blockID, actionID string) []string {
		raw := get(blockID, actionID)
		if raw == nil {
			return nil
		}
		opts, ok := raw["selected_options"].([]any)
		if !ok {
			return nil
		}
		out := make([]string, 0, len(opts))
		for _, o := range opts {
			m, ok := o.(map[string]any)
			if !ok {
				continue
			}
			if v, ok := m["value"].(string); ok && v != "" {
				out = append(out, v)
			}
		}
		return out
	}
	pickMultiValuesAndText := func(blockID, actionID string) (values []string, texts []string) {
		raw := get(blockID, actionID)
		if raw == nil {
			return nil, nil
		}
		opts, ok := raw["selected_options"].([]any)
		if !ok {
			return nil, nil
		}
		for _, o := range opts {
			m, ok := o.(map[string]any)
			if !ok {
				continue
			}
			v, _ := m["value"].(string)
			if v == "" {
				continue
			}
			values = append(values, v)
			label := ""
			if text, ok := m["text"].(map[string]any); ok {
				if s, ok := text["text"].(string); ok {
					label = s
				}
			}
			texts = append(texts, label)
		}
		return values, texts
	}

	pickPlainText := func(blockID, actionID string) string {
		raw := get(blockID, actionID)
		if raw == nil {
			return ""
		}
		v, _ := raw["value"].(string)
		return v
	}

	pickSelectedOptionText := func(blockID, actionID string) string {
		raw := get(blockID, actionID)
		if raw == nil {
			return ""
		}
		opt, ok := raw["selected_option"].(map[string]any)
		if !ok {
			return ""
		}
		text, ok := opt["text"].(map[string]any)
		if !ok {
			return ""
		}
		v, _ := text["text"].(string)
		return v
	}

	matchOpt := pickValue(subscribeMatchBlockID, subscribeMatchActionID, "selected_option")
	kindOpt := pickValue(subscribeKindBlockID, subscribeKindActionID, "selected_option")
	predicateOpt := pickValue(subscribePredicateBlockID, subscribePredicateActionID, "selected_option")

	rs := subscribeModalRenderState{
		OrgLinkID:        pickValue(subscribeOrgBlockID, subscribeOrgActionID, "selected_option"),
		Match:            matchOpt,
		TargetKind:       targetKindFromString(kindOpt),
		Predicate:        predicateOpt,
		LabelsRaw:        pickPlainText(subscribeLabelsBlockID, subscribeLabelsActionID),
		ExcludeLabelsRaw: pickPlainText(subscribeExcludeLabelsBlockID, subscribeExcludeLabelsActionID),
		Notif:            pickValue(subscribeNotifBlockID, subscribeNotifActionID, "selected_option"),
		Resources:        readResourceRenderStateFromValues(values, pickValue, pickMultiValues),
	}

	if matchOpt == matchOptionSpecific && targetKindNeedsApp(rs.TargetKind) {
		rs.AppID = pickValue(subscribeAppBlockID, subscribeAppActionID, "selected_option")
		rs.AppName = pickSelectedOptionText(subscribeAppBlockID, subscribeAppActionID)
	}

	// why: Read entities from whichever action_id matches the current kind.
	// Slack omits blocks the user can't see, so reading by the active
	// kind's action_id naturally drops any stale entries left over from
	// a previous render before the kind switched.
	if matchOpt == matchOptionSpecific {
		entityActionID := subscribeEntitiesActionIDForKind(rs.TargetKind)
		ids, names := pickMultiValuesAndText(subscribeEntitiesBlockID, entityActionID)
		rs.EntityIDs = ids
		rs.EntityNames = names
	}

	return rs
}

// why: readResourceRenderStateFromValues collects per-resource render state from
// view.state.values. Each enabled resource may contribute up to four
// blocks: opts (Enable), categories (Lifecycle / Approvals / Drift
// checkboxes), lifecycle (radio), subops (checkboxes). Slack omits hidden
// blocks, so absence means the user hasn't enabled the resource (or
// notif != specific).
//
// The categories checkbox group is the source of truth for which event
// streams are subscribed to:
//   - Lifecycle ticked → cfg.Outcome = the radio's value (default
//     Completion when the radio block is missing because we just re-ticked
//     Lifecycle and the previous radio state was dropped).
//   - Lifecycle un-ticked → cfg.Outcome = "none" (forces OutcomeNone in
//     buildSpecificEventsInterests).
//   - Approvals ticked → cfg.Approvals = true.
//   - Drift ticked → cfg.Drift = true (only meaningful for
//     components / sandboxes; the categories block omits the option for
//     other resources).
//   - Component health ticked → cfg.ComponentHealth = true (components
//     only); Install degraded ticked → cfg.InstallDegraded = true
//     (installs only). Same omit-the-option gating as Drift.
func readResourceRenderStateFromValues(
	values map[string]map[string]map[string]any,
	pickValue func(string, string, string) string,
	pickMultiValues func(string, string) []string,
) map[interests.ResourceKind]subscribeResourceCfg {
	out := make(map[interests.ResourceKind]subscribeResourceCfg, len(interests.AllResources))
	for _, kind := range interests.AllResources {
		optsBlock := subscribeResourceOptsBlockID(kind)
		categoriesBlock := subscribeResourceCategoriesBlockID(kind)
		lifecycleBlock := subscribeResourceLifecycleBlockID(kind)
		subOpsBlock := subscribeResourceSubOpsBlockID(kind)

		_, hasOpts := values[optsBlock]
		_, hasCategories := values[categoriesBlock]
		_, hasLifecycle := values[lifecycleBlock]
		_, hasSubOps := values[subOpsBlock]
		if !hasOpts && !hasCategories && !hasLifecycle && !hasSubOps {
			continue
		}

		cfg := subscribeResourceCfg{}
		for _, v := range pickMultiValues(optsBlock, subscribeResourceOptsActionID(kind)) {
			if v == resourceOptEnable {
				cfg.Enabled = true
			}
		}

		lifecycleOn := false
		for _, v := range pickMultiValues(categoriesBlock, subscribeResourceCategoriesActionID(kind)) {
			switch v {
			case categoryOptionLifecycle:
				lifecycleOn = true
			case categoryOptionApprovals:
				cfg.Approvals = true
			case categoryOptionDrift:
				cfg.Drift = true
			case categoryOptionComponentHealth:
				cfg.ComponentHealth = true
			case categoryOptionInstallDegraded:
				cfg.InstallDegraded = true
			}
		}

		if lifecycleOn {
			outcome := pickValue(lifecycleBlock, subscribeResourceLifecycleActionID(kind), "selected_option")
			if outcome == "" || outcome == outcomeOptionNone {
				outcome = outcomeOptionCompletion
			}
			cfg.Outcome = outcome
			for _, v := range pickMultiValues(subOpsBlock, subscribeResourceSubOpsActionID(kind)) {
				cfg.Ops = append(cfg.Ops, v)
			}
		} else {
			cfg.Outcome = outcomeOptionNone
		}
		out[kind] = cfg
	}
	return out
}

// why: handleSubscribeModalBlockSuggestion serves the install picker's
// external_select options. We re-derive the candidate install pool from the
// trusted TeamID in private_metadata (plus the user's currently-selected
// org-link) so the user can never request installs from another org.
//
// Returns the {options:[...]} body Slack expects.
func (s *service) handleSubscribeModalBlockSuggestion(ctx context.Context, payload slackInteractionPayload) map[string]any {
	empty := map[string]any{"options": []any{}}

	state, err := decodeSubscribeModalState(payload.View.PrivateMetadata)
	if err != nil {
		s.l.Warn("subscribe modal suggest: decode state failed", zap.Error(err))
		return empty
	}
	if state.TeamID != payload.Team.ID {
		s.l.Warn("subscribe modal suggest: team_id mismatch",
			zap.String("state_team_id", state.TeamID),
			zap.String("payload_team_id", payload.Team.ID))
		return empty
	}

	render := readSubscribeRenderStateFromPayload(payload)
	if render.OrgLinkID == "" {
		return empty
	}

	var link app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{
			ID:     render.OrgLinkID,
			TeamID: state.TeamID,
			Status: app.SlackOrgLinkStatusVerified,
		}).
		First(&link).Error; err != nil {
		s.l.Warn("subscribe modal suggest: org link not trusted", zap.Error(err))
		return empty
	}

	q := payload.Value
	const maxResults = 100

	tx := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Where(&app.Install{OrgID: link.OrgID}).
		Order("name ASC").
		Limit(maxResults)
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR id ILIKE ?", like, like)
	}
	var installs []app.Install
	if err := tx.Find(&installs).Error; err != nil {
		s.l.Warn("subscribe modal suggest: list installs failed", zap.Error(err))
		return empty
	}

	options := make([]any, 0, len(installs))
	for _, i := range installs {
		options = append(options, map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": truncatePlainText(entityPickerLabel(i.ID, i.Name), 75)},
			"value": i.ID,
		})
	}
	return map[string]any{"options": options}
}
