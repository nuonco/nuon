// Flat (resource, category) shape used by the redesigned picker UI. Maps onto
// the wire-format ResourceCfg fields below — kept here so both InterestsPicker
// (summary text + count) and InterestsModal (checklist) share one source of
// truth for what counts as a "selected event" and how to set/clear it.
//
// Backend outcome model (see ctl-api internal/pkg/interests/match.go):
//   - OutcomeNone:       no lifecycle events
//   - OutcomeAll / "":   every started + terminal event
//   - OutcomeCompletion: terminal events only
//   - OutcomeFailures:   failed/cancelled terminal events only
//
// The flattened UI collapses this to a single "lifecycle events" toggle per
// resource. Checking it writes Outcome="completion" (every termination);
// unchecking writes Outcome="none". The failures-only filter is intentionally
// dropped from this surface — power users can still craft it via API.

import type { ResourceCfg, ResourceKind } from './types'

export type EventCategory = 'lifecycle' | 'approvals' | 'drift'

export const RESOURCE_CATEGORIES: Record<ResourceKind, EventCategory[]> = {
  installs: ['lifecycle', 'approvals'],
  stacks: ['lifecycle'],
  components: ['lifecycle', 'approvals', 'drift'],
  sandboxes: ['lifecycle', 'approvals', 'drift'],
  install_configurations: ['lifecycle', 'approvals'],
  runners: ['lifecycle'],
  actions: ['lifecycle'],
}

export const CATEGORY_LABELS: Record<EventCategory, string> = {
  lifecycle: 'lifecycle events',
  approvals: 'approvals',
  drift: 'drift detected',
}

export const isCategoryOn = (
  cfg: ResourceCfg | undefined,
  cat: EventCategory
): boolean => {
  if (!cfg) return false
  switch (cat) {
    case 'lifecycle':
      return (cfg.outcome ?? 'all') !== 'none'
    case 'approvals':
      return !!cfg.approval_requests || !!cfg.approval_responses
    case 'drift':
      return !!cfg.drift_detected
  }
}

export const setCategoryOn = (
  cfg: ResourceCfg | undefined,
  cat: EventCategory,
  on: boolean
): ResourceCfg => {
  const base: ResourceCfg = cfg ?? {}
  switch (cat) {
    case 'lifecycle':
      return { ...base, outcome: on ? 'completion' : 'none' }
    case 'approvals':
      return { ...base, approval_requests: on, approval_responses: on }
    case 'drift':
      return { ...base, drift_detected: on }
  }
}

export const isResourceEmpty = (
  kind: ResourceKind,
  cfg: ResourceCfg | undefined
): boolean => RESOURCE_CATEGORIES[kind].every((c) => !isCategoryOn(cfg, c))
