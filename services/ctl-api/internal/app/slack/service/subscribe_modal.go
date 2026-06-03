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

// Subscribe-modal Block-Kit identifiers. Centralised so the open / re-render
// / view_submission handlers all key off the same ids — Slack rejects
// duplicate ids and silently drops mismatches when reading view.state.values.
const (
	subscribeModalCallbackID = "nuon_subscribe_modal"

	subscribeOrgBlockID  = "nuon_subscribe_org_block"
	subscribeOrgActionID = "nuon_subscribe_org"

	// Top-level "What should this match?" radio. Toggling it between
	// All / Specific dispatches a re-render that conditionally surfaces
	// the kind / predicate / entities-or-labels blocks.
	subscribeMatchBlockID  = "nuon_subscribe_match_block"
	subscribeMatchActionID = "nuon_subscribe_match"
	matchOptionAll         = "all"      // "Everything in this org" → Match=nil
	matchOptionSpecific    = "specific" // reveals kind+predicate

	// Resource-kind static_select. Only rendered when match=Specific.
	// Toggling re-renders to surface the predicate / entities-or-labels
	// blocks for the chosen kind.
	subscribeKindBlockID  = "nuon_subscribe_kind_block"
	subscribeKindActionID = "nuon_subscribe_kind"
	kindOptionInstalls    = "installs"
	kindOptionComponents  = "components"
	kindOptionActions     = "actions"

	// Predicate radio. Only rendered when match=Specific. Toggling
	// re-renders to surface either the multi_external_select or the
	// labels textinput.
	subscribePredicateBlockID  = "nuon_subscribe_predicate_block"
	subscribePredicateActionID = "nuon_subscribe_predicate"
	predicateOptionAny         = "any"      // empty TargetMatch{} for the chosen kind
	predicateOptionSpecific    = "specific" // multi_external_select → IDs
	predicateOptionLabels      = "labels"   // plain_text_input → Selector

	// App external_select. Only rendered when match=Specific and the
	// chosen kind scopes to an app (components / actions). Toggling it
	// dispatches a re-render so the entity picker below can re-source its
	// options against the chosen app.
	subscribeAppBlockID  = "nuon_subscribe_app_block"
	subscribeAppActionID = "nuon_subscribe_app"

	// Entity multi_external_select. Only one block at any time, but the
	// action_id varies by kind so the block_suggestion handler in Packet E
	// can dispatch on action_id without re-decoding state.
	subscribeEntitiesBlockID            = "nuon_subscribe_entities_block"
	subscribeEntitiesActionIDInstalls   = "nuon_subscribe_entities_installs"
	subscribeEntitiesActionIDComponents = "nuon_subscribe_entities_components"
	subscribeEntitiesActionIDActions    = "nuon_subscribe_entities_actions"

	// Labels selector textinput. Only rendered when predicate=Labels.
	// subscribeLabelsBlockID drives the positive (include) selector;
	// subscribeExcludeLabelsBlockID drives NotMatchLabels so users can
	// say "everything except env=stage" without enumerating every other
	// env value.
	subscribeLabelsBlockID         = "nuon_subscribe_labels_block"
	subscribeLabelsActionID        = "nuon_subscribe_labels"
	subscribeExcludeLabelsBlockID  = "nuon_subscribe_exclude_labels_block"
	subscribeExcludeLabelsActionID = "nuon_subscribe_exclude_labels"

	subscribeNotifBlockID  = "nuon_subscribe_notif_block"
	subscribeNotifActionID = "nuon_subscribe_notif"

	// Per-resource block / action ids are templated. They follow the
	// pattern nuon_subscribe_res_<kind>_(opts|categories|lifecycle|subops)[_block].
	subscribeResourceBlockIDPrefix = "nuon_subscribe_res_"

	notifOptionAll      = "all"
	notifOptionSpecific = "specific"

	// resourceOptEnable is the single option value used by the per-resource
	// Enable checkbox.
	resourceOptEnable = "enable"

	// Per-resource event-category checkbox values. The categories block is
	// the single point of progressive disclosure: ticking Lifecycle reveals
	// the lifecycle radio + sub-ops; ticking Approvals / Drift sets the
	// underlying booleans with no follow-up controls.
	categoryOptionLifecycle = "lifecycle"
	categoryOptionApprovals = "approvals"
	categoryOptionDrift     = "drift"

	// Per-resource lifecycle radio values; mirrors interests.Outcome.
	// outcomeOptionNone is no longer rendered in the radio — un-ticking the
	// Lifecycle category in the categories block is the new way to mute
	// lifecycle events. The constant is kept for the cfg.Outcome ↔
	// interests.Outcome mapping in buildSpecificEventsInterests and to
	// represent the "lifecycle off" state in subscribeResourceCfg.
	outcomeOptionNone       = "none"
	outcomeOptionAll        = "all"
	outcomeOptionCompletion = "completion"
	outcomeOptionFailures   = "failures"
)

// subscribeResourceOptsBlockID returns the input block id for the given
// resource's Enable checkbox.
func subscribeResourceOptsBlockID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_opts_block"
}

// subscribeResourceOptsActionID returns the action id for the resource's
// Enable checkbox element.
func subscribeResourceOptsActionID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_opts"
}

// subscribeResourceCategoriesBlockID returns the input block id for the
// resource's "Which event categories?" checkbox group (Lifecycle /
// Approvals / Drift detection). This block is the single point of
// progressive disclosure — ticking Lifecycle reveals the lifecycle radio
// + sub-ops blocks; ticking Approvals / Drift sets the underlying booleans
// with no follow-up controls.
func subscribeResourceCategoriesBlockID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_categories_block"
}

func subscribeResourceCategoriesActionID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_categories"
}

// subscribeResourceLifecycleBlockID returns the input block id for the
// resource's lifecycle radio (All / On completion / On failures). Only
// rendered when the Lifecycle category is ticked in the categories block
// — the radio no longer carries an "Off" option in the new design.
func subscribeResourceLifecycleBlockID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_lifecycle_block"
}

func subscribeResourceLifecycleActionID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_lifecycle"
}

// subscribeResourceSubOpsBlockID returns the input block id for the
// resource's lifecycle sub-op checkbox group. Rendered inline below the
// lifecycle radio whenever the Lifecycle category is ticked.
func subscribeResourceSubOpsBlockID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_subops_block"
}

func subscribeResourceSubOpsActionID(k interests.ResourceKind) string {
	return subscribeResourceBlockIDPrefix + string(k) + "_subops"
}

// subscribeModalState is round-tripped via the modal's private_metadata so
// re-renders / view_submission don't have to refetch slash-command context.
// The TeamID is the originally-signed team id from the slash payload — we
// re-verify all selected ids (org_link, install) against it on every action.
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

// subscribeResourceCfg captures the per-resource modal state for the
// "Specific events" expanded view. Mirrors a single ResourceCfg row in
// interests.Interests.Resources, plus an Enabled bit so the user can toggle
// individual resources off without losing their other knob selections (we
// preserve these per-resource selections across re-renders by initialising
// the checkbox group from this struct).
type subscribeResourceCfg struct {
	Enabled   bool
	Ops       []string // sub-op slugs (matches interests.SubOps[kind])
	Approvals bool     // collapses approval_requests + approval_responses
	Drift     bool     // only meaningful for components / sandboxes
	Outcome   string   // outcomeOptionNone | outcomeOptionAll | outcomeOptionCompletion | outcomeOptionFailures
}

// subscribeModalRenderState captures the user's currently-selected modal
// values so re-renders preserve them. Defaults baked into buildSubscribeModalView
// when zero values come in.
type subscribeModalRenderState struct {
	OrgLinkID string

	// Match selects between the two top-level routing modes. matchOptionAll
	// emits a SubscriptionMatch=nil row (org-wide). matchOptionSpecific
	// reveals the kind / predicate / entities-or-labels triplet below.
	Match string // matchOptionAll | matchOptionSpecific

	// TargetKind names the resource taxonomy the predicate applies to.
	// Only meaningful when Match==matchOptionSpecific. Unset defaults to
	// installs on the first render of the Specific path.
	TargetKind labels.TargetKind

	// AppID scopes the entity picker when TargetKind ∈ {Components,
	// Actions} and Predicate=Specific. Components and actions belong to
	// an app, so the user picks an app first and the entity picker is
	// then filtered by app_id. AppName is the display label for the
	// currently-selected app, carried across re-renders so the
	// external_select can preselect itself without round-tripping
	// through the block_suggestion handler. Ignored for installs and
	// for the labels predicate (labels span apps by design).
	AppID   string
	AppName string

	// Predicate is the per-kind matcher: Any (empty TargetMatch{}),
	// Specific (multi_external_select → IDs), or Labels (text → Selector).
	// Only meaningful when Match==matchOptionSpecific.
	Predicate string // predicateOption{Any,Specific,Labels}

	// EntityIDs holds the multi_external_select's chosen ids when
	// Predicate==predicateOptionSpecific. EntityNames is parallel to
	// EntityIDs and carries the display label for each id (used to seed
	// initial_options across re-renders without a round-trip via the
	// block_suggestion handler).
	EntityIDs   []string
	EntityNames []string

	// LabelsRaw is the verbatim text the user typed in the labels
	// selector textinput when Predicate==predicateOptionLabels. Parsed
	// lazily via labels.ParseLabelsQuery on submission and on preview.
	LabelsRaw string

	// ExcludeLabelsRaw is the verbatim text from the exclusion selector
	// textinput (same predicate). Parsed into Selector.NotMatchLabels so
	// "everything except env=stage" works without enumerating positives.
	// Either LabelsRaw or ExcludeLabelsRaw (or both) may be empty; at
	// least one must be non-empty for the Labels predicate to validate.
	ExcludeLabelsRaw string

	Notif string // notifOptionAll | notifOptionSpecific

	// Resources is populated only when Notif==notifOptionSpecific. When
	// nil/empty on initial render, buildSubscribeModalView seeds it from
	// interests.Default() so users land on the same baseline the dashboard
	// uses (4 of 6 resources enabled, completion outcome, approvals on,
	// drift on for components+sandboxes).
	Resources map[interests.ResourceKind]subscribeResourceCfg
}

// subscribePreview is the live "currently matches N" payload rendered
// inline below the entity / labels picker when match=Specific and
// predicate≠Any. The preview block is only rendered when this is non-nil
// and Count > 0 or Err != nil. Caller (openSubscribeModalForSlash /
// rerenderSubscribeModal) computes it via service.previewMatched; the
// modal's pure-rendering path tests pass nil here so no DB is needed.
type subscribePreview struct {
	Kind  labels.TargetKind
	Count int
	Names []string
	Err   error
}

// buildSubscribeModalView renders the subscribe modal Block-Kit JSON. Slack
// modals are limited (~25 blocks, no inline conditionals), so conditional
// affordances are realised by re-rendering the entire modal on every
// block_actions tick.
//
// links is the list of verified org-links for the workspace; an empty list
// is rejected before we ever open a modal. preview is optional — pass nil
// to suppress the live "currently matches N" context block (tests do this).
func buildSubscribeModalView(
	state subscribeModalState,
	links []app.SlackOrgLink,
	render subscribeModalRenderState,
	preview *subscribePreview,
) (map[string]any, error) {
	if len(links) == 0 {
		return nil, errors.New("subscribe modal: no verified org links")
	}

	// Default the org-link selection to the first link if the caller didn't
	// pre-pick one. Same for the top-level Match radio (org-wide is the
	// dashboard's default).
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

	// Org picker. Always rendered. With 1 link Slack still shows a select
	// (preselected); with N links the user picks the source org.
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
		"type":     "input",
		"block_id": subscribeOrgBlockID,
		"label":    map[string]any{"type": "plain_text", "text": "Nuon org"},
		"element":  orgSelect,
		// dispatch_action so a re-render carries the user's pick forward
		// (org changes don't re-render anything else right now but we
		// keep the option on so future per-org affordances can plug in).
		"dispatch_action": false,
	})

	// Top-level "What should this match?" radio. dispatch_action=true so
	// toggling it triggers a block_actions re-render that conditionally
	// surfaces the kind / predicate / entities-or-labels triplet below.
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

	// Specific path — kind picker + predicate + (entities | labels) +
	// preview. Only rendered when Match==Specific.
	if render.Match == matchOptionSpecific {
		// Resource-kind static_select. dispatch_action=true so a
		// kind change re-renders to surface the right entity picker
		// action_id (and clears stale entity / labels selections —
		// see clearStaleSpecificFields in the block_actions path).
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
		kindOptions := []any{kindOptInstalls, kindOptComponents, kindOptActions}
		kindInitial := kindOptInstalls
		switch render.TargetKind {
		case labels.TargetKindComponents:
			kindInitial = kindOptComponents
		case labels.TargetKindActions:
			kindInitial = kindOptActions
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

		// Predicate radio. dispatch_action=true so a predicate change
		// re-renders to surface either the multi_external_select or
		// the labels textinput (and re-runs the preview).
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
			// Components and actions belong to an app, so we
			// surface an app picker first; the entity picker only
			// renders once an app has been chosen. Installs are
			// org-scoped so they skip the app gate.
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
					"type":     "input",
					"block_id": subscribeAppBlockID,
					"label":    map[string]any{"type": "plain_text", "text": "App"},
					"element":  appSelect,
					// dispatch_action so changing the app re-renders
					// the entity picker against the new app's
					// components / actions.
					"dispatch_action": true,
				})
			}

			// Entity multi_external_select — action_id varies by
			// kind so the block_suggestion handler can dispatch
			// without re-decoding state. For app-scoped kinds we
			// only render the entity picker once an app has been
			// chosen; otherwise the picker would have nothing to
			// load.
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
					"type":     "input",
					"block_id": subscribeEntitiesBlockID,
					"label":    map[string]any{"type": "plain_text", "text": titleCase(targetKindLabel(render.TargetKind, true))},
					"element":  entSelect,
					// dispatch_action so the preview re-runs as the user
					// adds / removes entities (Slack only emits
					// block_actions for inputs with dispatch_action set).
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
				"type":     "input",
				"block_id": subscribeLabelsBlockID,
				"label":    map[string]any{"type": "plain_text", "text": "Include labels"},
				"hint":     map[string]any{"type": "plain_text", "text": "Leave empty to include everything (combine with Exclude labels below)."},
				"optional": true,
				"element":  labelsInput,
				// dispatch_action so the preview re-runs on enter; we
				// also re-run on every other dispatch_action that
				// re-renders the modal.
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

		// Live preview block — context section under the picker. Only
		// rendered for non-Any predicates; Any is "match every entity
		// of this kind" by definition and the preview text would just
		// be redundant with the predicate label.
		if render.Predicate != predicateOptionAny && preview != nil {
			blocks = append(blocks, buildPreviewContextBlock(render.TargetKind, preview))
		}
	}

	// Notification radio. Mute is gone with the install pin — every
	// remaining sub fires according to its own per-resource Interests.
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

	// Per-resource event-filter blocks — only when notification=specific.
	// Each enabled resource expands progressively:
	//   1. Enable checkbox (always rendered)
	//   2. "Which event categories?" checkbox group (Lifecycle / Approvals
	//      / +Drift detection on components+sandboxes) — rendered when
	//      Enable is ticked.
	//   3. Lifecycle radio (All / On completion / On failures) — rendered
	//      only when the Lifecycle category is ticked.
	//   4. Sub-op checkbox group — rendered inline beneath the lifecycle
	//      radio (Slack lacks accordions, so always-visible-when-lifecycle-on
	//      is the accepted compromise vs the dashboard's secondary
	//      disclosure).
	// Approvals and Drift have no follow-up controls — ticking the
	// category in the categories block is the only knob.
	//
	// On the first render after switching to "Specific events" the render
	// state is empty; we seed each resource from interests.Default() so
	// users land on the same baseline the dashboard's per-resource picker
	// uses.
	if render.Notif == notifOptionSpecific {
		seeded := seedResourcesForRender(render.Resources)
		blocks = append(blocks, map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": "*Per-resource filters*\nEnable the resources you care about, then pick which event categories (lifecycle, approvals, drift detection) you want notifications for.",
			},
		})
		for _, kind := range interests.AllResources {
			cfg := seeded[kind]
			blocks = append(blocks, map[string]any{"type": "divider"})
			blocks = append(blocks, buildResourceOptsBlock(kind, cfg))
			if cfg.Enabled {
				blocks = append(blocks, buildResourceCategoriesBlock(kind, cfg))
				// Lifecycle details only appear when the Lifecycle
				// category is ticked. Empty Outcome is treated as on
				// (the seed defaults to OutcomeCompletion for
				// pre-enabled resources).
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

// resourceKindLabel returns a human-readable label for a ResourceKind. Slack
// modal text is plain-text only; spaces are fine.
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

// subOpLabel returns a human-readable label for a sub-op slug. Mirrors the
// dashboard's labelForSubOp helper so the two surfaces feel consistent.
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
	case "run":
		return "Run"
	case "version_active":
		return "Version active"
	default:
		return op
	}
}

// seedResourcesForRender returns the per-resource render state to use when
// drawing the "Specific events" expanded view. If the caller passed a
// non-empty map (typical re-render via view.state.values), it's returned
// as-is. Otherwise we materialise interests.Default() as the baseline so
// users land on the same starting shape the dashboard variant ships.
func seedResourcesForRender(in map[interests.ResourceKind]subscribeResourceCfg) map[interests.ResourceKind]subscribeResourceCfg {
	if len(in) > 0 {
		return in
	}
	defaults := interests.Default().Resources
	out := make(map[interests.ResourceKind]subscribeResourceCfg, len(interests.AllResources))
	for _, kind := range interests.AllResources {
		cfg, ok := defaults[kind]
		if !ok {
			// Resources not in Default() (runners, actions) start
			// disabled; users opt in explicitly. Seed Outcome to
			// Completion so that re-enabling pre-fills the lifecycle
			// radio with the same baseline the dashboard uses.
			out[kind] = subscribeResourceCfg{Outcome: outcomeOptionCompletion}
			continue
		}
		out[kind] = subscribeResourceCfg{
			Enabled:   true,
			Ops:       append([]string(nil), cfg.Ops...),
			Approvals: cfg.ApprovalRequests || cfg.ApprovalResponses,
			Drift:     cfg.DriftDetected,
			Outcome:   outcomeFromInterests(cfg.Outcome),
		}
	}
	return out
}

// outcomeFromInterests projects an interests.Outcome into the radio value
// vocabulary used by the modal. Empty / unknown values fall back to "all"
// (matches interests.match's treatment of an unset outcome).
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

// outcomeToInterests is the inverse of outcomeFromInterests. Unknown values
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

// outcomeRadioOptions returns the three lifecycle radio options surfaced
// in the modal. The "Off" option (OutcomeNone) is no longer rendered —
// un-ticking the Lifecycle category in the categories block is the new
// way to mute lifecycle events while leaving approvals and drift
// independently gated.
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

// buildResourceOptsBlock renders the per-resource Enable checkbox. Toggling
// it dispatches block_actions, and the modal re-renders to reveal (or hide)
// the lifecycle / approvals / drift sub-blocks.
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

// buildResourceLifecycleBlock renders the per-resource lifecycle radio
// (All events / On completion / On failures). Only rendered when the
// Lifecycle category is ticked in the categories block — un-ticking the
// category is the new "Off". An empty or "none" Outcome defaults to
// Completion to match the dashboard baseline (and to ensure re-ticking
// Lifecycle restores a sensible value when the previous radio state was
// dropped from view.state.values during the hidden render).
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

// buildResourceCategoriesBlock renders the per-resource "Which event
// categories?" checkbox group — the single point of progressive
// disclosure. Options are Lifecycle, Approvals, and (for components +
// sandboxes) Drift detection. dispatch_action is set so ticking Lifecycle
// re-renders the modal to reveal the lifecycle radio + sub-ops blocks
// (and un-ticking it hides them).
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
	// Lifecycle is on whenever Outcome is anything other than "none".
	// Empty Outcome counts as on (matches the matcher's treatment of an
	// unset outcome).
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

// buildResourceSubOpsBlock renders the per-resource lifecycle sub-op
// checkbox group. Only meaningful when the lifecycle radio is not Off.
// Leaving every checkbox unchecked is treated as "every sub-op" by the
// matcher (Ops is persisted as nil/empty).
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

// entityPickerLabel formats an entity (install / component / action) for
// display in a multi_external_select. We prefer "Name (id)" when both are
// known; falls back to bare id.
func entityPickerLabel(id, name string) string {
	if name == "" {
		return id
	}
	return fmt.Sprintf("%s (%s)", name, id)
}

// targetKindLabel returns a human-readable noun for the given TargetKind.
// plural=true returns the plural form (used in field labels like "Search
// installs"); plural=false returns the singular ("install").
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
	default:
		// Default to installs so an unset TargetKind still produces
		// useful labels during the first render of the Specific path.
		if plural {
			return "installs"
		}
		return "install"
	}
}

// titleCase upper-cases the first byte of s. Used to turn "installs" →
// "Installs" for the entity picker block label.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// targetKindNeedsApp reports whether the entity picker for k must be
// scoped to a specific app. Components and actions belong to an app, so
// the modal surfaces an app picker first; installs are listed org-wide.
func targetKindNeedsApp(k labels.TargetKind) bool {
	switch k {
	case labels.TargetKindComponents, labels.TargetKindActions:
		return true
	default:
		return false
	}
}

// subscribeEntitiesActionIDForKind dispatches the entity-picker action_id
// on the chosen TargetKind. The block_suggestion handlers (install lives
// here, components+actions land in Packet E) key off these constants so
// they don't have to re-decode the modal's render state to know what kind
// of entity to fetch.
func subscribeEntitiesActionIDForKind(k labels.TargetKind) string {
	switch k {
	case labels.TargetKindComponents:
		return subscribeEntitiesActionIDComponents
	case labels.TargetKindActions:
		return subscribeEntitiesActionIDActions
	default:
		return subscribeEntitiesActionIDInstalls
	}
}

// targetKindFromString maps a kindOption* value back to a labels.TargetKind.
// Unknown values fall through to TargetKindInstalls so a tampered payload
// can't smuggle an unknown kind into the matcher.
func targetKindFromString(s string) labels.TargetKind {
	switch s {
	case kindOptionComponents:
		return labels.TargetKindComponents
	case kindOptionActions:
		return labels.TargetKindActions
	default:
		return labels.TargetKindInstalls
	}
}

// buildPreviewContextBlock renders the live "currently matches N: a, b, c"
// context block surfaced under the entity / labels picker. Slack's context
// block accepts mrkdwn, so we use bold for the count and a comma-joined
// list of the first 6 names + "+N more" when truncated.
//
// Three rendered shapes:
//   - Err != nil → "_Couldn't compute preview: <err>_"
//   - Count == 0 → ":warning: No <kind> currently match this filter."
//   - Count > 0  → "Currently matches *N* <kind>: a, b, c, …"
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

// truncatePlainText caps a label to Slack's 75-byte plain_text limit.
// The ellipsis is U+2026 (3 UTF-8 bytes), so we reserve 3 bytes from max
// — using "max-1" would over-shoot the cap by 2 bytes and Slack rejects
// any plain_text option whose text exceeds 75 bytes.
func truncatePlainText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	const ellipsis = "…"
	if max <= len(ellipsis) {
		// Not enough room for a meaningful suffix; hard-truncate at a
		// rune boundary so we never emit an invalid UTF-8 byte sequence.
		return safeBytePrefix(s, max)
	}
	return safeBytePrefix(s, max-len(ellipsis)) + ellipsis
}

// safeBytePrefix returns the longest prefix of s whose UTF-8 length is
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
	// Walk runes until adding the next one would exceed n bytes.
	out := 0
	for i := range s {
		if i > n {
			break
		}
		out = i
	}
	return s[:out]
}

// openSubscribeModalForSlash opens a fresh subscribe modal in response to a
// /nuon subscribe slash command. Caller has already verified the workspace
// has an active install + ≥1 verified org-link. preselect lets the caller
// pre-populate the modal — pass the zero value to use the modal's empty
// defaults, or pass a state derived from an existing channel subscription
// (see preselectSubscribeRenderStateForChannel) so the user doesn't lose
// context when re-running /nuon subscribe in a channel that already has a
// subscription configured.
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

	// Drop a preselected OrgLinkID that no longer points at a verified
	// link (link revoked since the existing sub was created); the modal
	// will fall back to its default-first-link behaviour.
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

	// Compute live preview if the preselect already specifies a non-Any
	// predicate. Skipped silently on error — the preview block degrades
	// to a "couldn't compute preview" line.
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

// preselectSubscribeRenderStateForChannel derives a subscribeModalRenderState
// from the most recently updated SlackChannelSubscription for the given
// (teamID, channelID), so re-running /nuon subscribe in a channel that
// already has a subscription doesn't lose the user's prior selections.
//
// Returns the zero value when there is no existing subscription (or any
// lookup error) — the caller falls through to the modal's empty defaults.
// We intentionally surface only one row even when the channel has multiple
// subscriptions (e.g. an org-wide row plus install-scoped overrides): the
// user can still flip the scope / install / notification radios in-modal,
// but their per-resource event filters carry over instead of being reset.
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

// renderStateFromSubscription projects a stored SlackChannelSubscription back
// into the modal's subscribeModalRenderState shape. Inverse of the
// match/notif/per-resource translations performed by
// handleSubscribeModalSubmission + buildSpecificEventsInterests.
//
// Projection rules for sub.Match:
//   - nil → Match=All
//   - first non-nil *TargetMatch in (Installs, Components, Actions):
//   - if t.IDs non-empty: Predicate=Specific, EntityIDs=t.IDs,
//     EntityNames batch-loaded from the matching table
//   - else if t.Selector!=nil: Predicate=Labels, LabelsRaw=join of
//     MatchLabels back to "k=v,k=*"
//   - else (empty TargetMatch{}): Predicate=Any
//
// We surface only the first populated kind even though the schema permits
// multi-kind in one row — the v1 modal can't render that anyway, and the
// user can still flip the kind picker in-modal to inspect the others.
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
		}
		rs.TargetKind = kind
		switch {
		case len(tm.IDs) > 0:
			rs.Predicate = predicateOptionSpecific
			rs.EntityIDs = append([]string(nil), tm.IDs...)
			rs.EntityNames = s.lookupEntityNames(ctx, sub.OrgID, kind, tm.IDs)
			// Components and actions belong to an app — re-derive
			// the owning app from the first id so the app picker
			// preselects to the same app the entity ids came from.
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
		// Existing-row Interests defaults to AllEvents=true at create
		// time; a zero-valued Interests on disk almost certainly means
		// "never explicitly configured". Fall back to All so the modal
		// doesn't render a fully-disabled per-resource picker.
		rs.Notif = notifOptionAll
	} else {
		rs.Notif = notifOptionSpecific
		rs.Resources = renderResourcesFromInterests(sub.Interests)
	}
	return rs
}

// describeMatch renders a SlackChannelSubscription.Match into a short,
// human-readable scope tag for the unsubscribe modal and /nuon status. The
// vocabulary mirrors the subscribe modal's UI labels so users see the same
// noun in both directions:
//
//   - nil Match (or all-nil targets) → "everything in org"
//   - first populated kind:
//   - Selector → "by labels: env=prod, owner=*"
//   - IDs → "specific {kind}: a, b, c, +N more"
//   - empty TargetMatch{} → "any {kind}"
//
// We surface only the first populated kind because the v1 modal only ever
// writes one — multi-kind rows are a CLI-only construct today.
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

// labelsToQueryString joins a labels.Labels map back into the
// k=v,k=*-grammar text the labels textinput consumes. Keys are sorted so
// the rendered string is deterministic across re-renders. Wildcard values
// ("*") collapse to "k" with no value, matching the bare-key syntax
// labels.ParseLabelsQuery accepts on input.
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

// lookupEntityNames batch-loads display names for a list of entity ids of
// the given TargetKind, returning a slice parallel to the input. Missing
// rows produce empty-string entries (the picker falls back to bare id via
// entityPickerLabel). Errors are swallowed and logged — the modal still
// renders, just without pre-populated names.
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

// lookupOwningApp returns (id, name) of the app that owns the given
// component / action id. Used by renderStateFromSubscription to
// preselect the app picker when re-opening the modal on an existing
// subscription. Errors and missing rows degrade silently to empty
// strings; the modal will then prompt the user to pick an app afresh.
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

// tableForTargetKind maps a labels.TargetKind to the GORM table name we
// query for preview / name lookups. Returns "" for an unknown kind.
func tableForTargetKind(kind labels.TargetKind) string {
	switch kind {
	case labels.TargetKindInstalls:
		return "installs"
	case labels.TargetKindComponents:
		return "components"
	case labels.TargetKindActions:
		return "action_workflows"
	default:
		return ""
	}
}

// renderResourcesFromInterests converts a stored interests.Interests config
// into the per-resource render-state map the modal expects. Inverse of
// buildSpecificEventsInterests: every resource in the canonical list is
// represented (so the modal renders deterministically), but only resources
// present in the stored Interests are flagged Enabled.
func renderResourcesFromInterests(in interests.Interests) map[interests.ResourceKind]subscribeResourceCfg {
	out := make(map[interests.ResourceKind]subscribeResourceCfg, len(interests.AllResources))
	for _, kind := range interests.AllResources {
		if cfg, ok := in.Resources[kind]; ok {
			out[kind] = subscribeResourceCfg{
				Enabled:   true,
				Ops:       append([]string(nil), cfg.Ops...),
				Approvals: cfg.ApprovalRequests || cfg.ApprovalResponses,
				Drift:     cfg.DriftDetected,
				Outcome:   outcomeFromInterests(cfg.Outcome),
			}
			continue
		}
		// Resource isn't in the stored config — render it as a
		// disabled row pre-seeded with the same Completion baseline
		// the dashboard uses so re-enabling it lands on a sensible
		// default radio value.
		out[kind] = subscribeResourceCfg{Outcome: outcomeOptionCompletion}
	}
	return out
}

// lookupActiveInstallForTeam fetches the active SlackInstallation for a team
// id. Used by the modal flows to obtain the bot token for views.open /
// views.update / external_select option lookup.
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

// rerenderSubscribeModal is invoked by the block_actions handler when the
// scope or notification radio changes. Re-derives the org-link list from
// the trusted TeamID in private_metadata and pushes a fresh view via
// views.update.
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
		// Someone tampered with private_metadata between renders — refuse.
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

// handleSubscribeModalBlockActions handles user pivots inside the subscribe
// modal — specifically scope/notification radio changes that flip which
// blocks are visible. Returns 200 with no body to ack the interaction; the
// re-render itself happens via the views.update call inside.
func (s *service) handleSubscribeModalBlockActions(ctx context.Context, payload slackInteractionPayload) {
	render := readSubscribeRenderStateFromPayload(payload)

	// Switching the app means the prior entity selections refer to
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

// handleSubscribeModalSubmission resolves the user's selections, validates
// trust (org_link belongs to TeamID; install belongs to link.OrgID for
// install-scoped subs), and creates the SlackChannelSubscription via
// upsertModalSubscription.
//
// Returns Slack's response_action JSON envelope:
//
//   - {} (empty 200) → close the modal on success
//   - {"response_action":"errors","errors":{"<block_id>":"..."}} → display
//     inline errors and keep the modal open
//
// Caller is responsible for writing the JSON to ctx — we return the body so
// it's testable.
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

	// Trust-bind the chosen org link to the signed TeamID. If the user
	// somehow posted a link id from another workspace it isn't found.
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

	// Compose the SubscriptionMatch from the modal selections. errBlock
	// is the offending block_id when validation fails, surfaced as an
	// inline Slack errors envelope.
	match, errBlock, errMsg := buildSubscriptionMatchFromRender(render)
	if errMsg != "" {
		return modalErr(errBlock, errMsg)
	}

	// Anti-tampering: when Predicate=Specific, all chosen entity IDs
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

	// Build the Interests config from the per-resource render state.
	// notifOptionAll → AllEvents=true; notifOptionSpecific →
	// per-resource picker output. The legacy Mute branch is gone.
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

	// Empty 200 closes the modal.
	return map[string]any{}
}

// buildSubscriptionMatchFromRender composes a *labels.SubscriptionMatch
// from the modal's render state. Returns (nil, "", "") for the org-wide
// case (Match=All). On validation failure returns (nil, blockID, message)
// so the caller can surface a Slack errors envelope.
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
	default: // predicateOptionAny
		tm = &labels.TargetMatch{}
	}

	out := &labels.SubscriptionMatch{}
	switch kind {
	case labels.TargetKindComponents:
		out.Components = tm
	case labels.TargetKindActions:
		out.Actions = tm
	default:
		out.Installs = tm
	}
	if err := out.Validate(); err != nil {
		// Selector parse already covers most reachable failure modes;
		// surface generic message keyed on the entity block as a
		// last-resort pointer.
		return nil, subscribeEntitiesBlockID, "Invalid filter: " + err.Error()
	}
	return out, "", ""
}

// validateMatchEntityIDs ensures every Specific id in the match belongs to
// orgID. Trust-binding is otherwise transitive via the org_link, but the
// view_submission body is user-controlled. For component / action kinds
// the picker is further scoped to appID; we re-apply that filter so a
// tampered payload can't reference entities from another app in the same
// org.
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
	return "", ""
}

// buildSpecificEventsInterests materialises a per-resource Interests config
// from the modal's per-resource render state. Only resources with
// Enabled==true contribute a row; everything else is dropped (the matcher
// treats absence as "not interested"). Approval booleans are mirrored to
// both ApprovalRequests + ApprovalResponses, mirroring the dashboard slack
// variant's collapse choice.
//
// Sub-op slugs we don't recognise (eg. an attacker mutating view.state)
// are filtered out — only slugs declared in interests.SubOps[kind] survive.
// Drift only persists for resources where SupportsDriftDetected returns
// true, so an unsupported drift selection from a tampered payload is a
// no-op rather than a silently-dropped invariant.
func buildSpecificEventsInterests(in map[interests.ResourceKind]subscribeResourceCfg) interests.Interests {
	out := interests.Interests{
		Resources: make(map[interests.ResourceKind]interests.ResourceCfg),
	}
	for _, kind := range interests.AllResources {
		cfg, ok := in[kind]
		if !ok || !cfg.Enabled {
			continue
		}

		// Filter sub-ops to the canonical vocabulary so a tampered
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
		out.Resources[kind] = rc
	}
	return out
}

// upsertModalSubscription writes exactly one SlackChannelSubscription row
// per modal submission. The unique index
// idx_slack_channel_subs_team_channel_link_match collapses
// (team, channel, link, match_canonical) so re-submitting the same modal
// updates Interests + ChannelName in place rather than duplicating the row.
//
// match==nil persists as SQL NULL (org-wide) — see
// labels.SubscriptionMatch.Value.
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
	// The slack-side flow has no account context, so satisfy the
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

// maybePreview computes the live preview for a given render state. Returns
// nil when there is nothing to preview (Match=All, Predicate=Any, or no
// resolvable orgID). Errors are folded into the returned struct so the
// preview block degrades gracefully instead of failing the whole render.
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

// previewMatched runs the live "currently matches N <kind>" query against
// the chosen kind's table. Returns (0, nil, nil) for the "match nothing"
// case (nil/empty TargetMatch); the caller skips the preview block in
// that case via maybePreview.
//
// Capped at 7 rows so the preview surfaces "+N more" when truncated.
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

	// First the total count, then the truncated name list.
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

// modalErr builds a response_action=errors envelope so the modal stays open
// with an inline error attached to the offending block.
func modalErr(blockID, message string) map[string]any {
	return map[string]any{
		"response_action": "errors",
		"errors": map[string]string{
			blockID: message,
		},
	}
}

// readSubscribeRenderStateFromPayload extracts the user's currently-selected
// modal values from view.state.values. Slack reports state for every input
// block in the view, even ones the user hasn't touched (when an
// initial_option is set). Resilient to missing keys (returns zero-string for
// any block not in the payload).
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
	// pickMultiValuesAndText reads a multi-select block's selected_options,
	// returning parallel slices of values and their plain_text labels. Used
	// by the install picker so we can round-trip both id and name across
	// re-renders without re-querying the block_suggestion source.
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

	// pickPlainText reads a plain_text_input block's "value" string.
	// Slack's wire shape for plain_text_input is {type, value} (flat,
	// no selected_option wrapper) so it doesn't fit pickValue's
	// option-pluck shape.
	pickPlainText := func(blockID, actionID string) string {
		raw := get(blockID, actionID)
		if raw == nil {
			return ""
		}
		v, _ := raw["value"].(string)
		return v
	}

	// pickSelectedOptionText reads the plain_text label of a single
	// selected_option, alongside its value. external_select / static_select
	// payloads carry the option's label so we can round-trip a display
	// name across re-renders without re-querying the suggestion source.
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

	// Read app picker selection. Slack only emits this block when it
	// was rendered, so absence on the wire safely zeroes AppID/AppName
	// for kinds that don't gate on an app (installs, labels predicate).
	if matchOpt == matchOptionSpecific && targetKindNeedsApp(rs.TargetKind) {
		rs.AppID = pickValue(subscribeAppBlockID, subscribeAppActionID, "selected_option")
		rs.AppName = pickSelectedOptionText(subscribeAppBlockID, subscribeAppActionID)
	}

	// Read entities from whichever action_id matches the current kind.
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

// readResourceRenderStateFromValues collects per-resource render state from
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

		// Read the categories checkbox group. If the categories block
		// isn't in the payload (resource not yet enabled in this
		// render), all categories default to off and Outcome defaults
		// to "none" so the resource is a no-op until the user ticks
		// something.
		lifecycleOn := false
		for _, v := range pickMultiValues(categoriesBlock, subscribeResourceCategoriesActionID(kind)) {
			switch v {
			case categoryOptionLifecycle:
				lifecycleOn = true
			case categoryOptionApprovals:
				cfg.Approvals = true
			case categoryOptionDrift:
				cfg.Drift = true
			}
		}

		if lifecycleOn {
			// Lifecycle is on -> read the radio. Hidden radio (no
			// payload) defaults to Completion to match the
			// dashboard baseline and the seed's default.
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

// handleSubscribeModalBlockSuggestion serves the install picker's
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

	// Re-verify the org link is trusted for this workspace (TeamID-bound).
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

	// Filter installs by the user-typed query (case-insensitive substring
	// against name and id). Cap at 100 — Slack's external_select max.
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
