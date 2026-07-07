// TypeScript mirror of the Go contract in
// services/ctl-api/internal/pkg/interests/types.go.
//
// The wire format uses snake_case, exactly as the Go struct tags emit it.
// Backend types are stamped `swaggertype:"object"`, so the auto-generated SDK
// surface for `interests` is a generic object — these hand-written types are
// the canonical shape everything client-side reads / writes.

export type ResourceKind =
  | 'installs'
  | 'stacks'
  | 'components'
  | 'sandboxes'
  | 'install_configurations'
  | 'runners'
  | 'actions'
  | 'app_branches'

export type Outcome = 'none' | 'all' | 'completion' | 'failures'

export interface ResourceCfg {
  ops?: string[]
  outcome?: Outcome
  approval_requests?: boolean
  approval_responses?: boolean
  // drift_detected fires only when drift is actually detected during a drift
  // scan (not for clean scans). Independent of `outcome`. Only meaningful for
  // components and sandboxes (see RESOURCES_WITH_DRIFT_DETECTED); harmless
  // but never matches on others. Drift workflow lifecycle events themselves
  // are unconditionally suppressed by the matcher — this flag is the only
  // knob that surfaces drift to subscribers.
  drift_detected?: boolean
  // config_synced fires when app config is synced to an app branch,
  // independent of lifecycle outcome. Only meaningful for app_branches
  // (see SupportsConfigSynced); harmless but never matches on others.
  config_synced?: boolean
}

export interface Interests {
  all_events?: boolean
  resources?: Partial<Record<ResourceKind, ResourceCfg>>
}

// Canonical, ordered list of resource kinds. Mirrors Go AllResources.
// UI rows render in this order; defaults() walks it.
export const ALL_RESOURCES: ResourceKind[] = [
  'installs',
  'stacks',
  'components',
  'sandboxes',
  'install_configurations',
  'runners',
  'actions',
  'app_branches',
]

// Human label per resource. Sentence-case.
export const RESOURCE_LABELS: Record<ResourceKind, string> = {
  installs: 'Installs',
  stacks: 'Stacks',
  components: 'Components',
  sandboxes: 'Sandboxes',
  install_configurations: 'Install configurations',
  runners: 'Runners',
  actions: 'Actions',
  app_branches: 'App branches',
}

export const RESOURCE_DESCRIPTIONS: Record<ResourceKind, string> = {
  installs: 'Install provision, deprovision, reprovision lifecycle.',
  stacks: 'Install stack lifecycle, including when a stack version becomes active.',
  components: 'Per-component deploy and teardown. Toggle drift detected to be notified when drift is found.',
  sandboxes: 'Sandbox provision, reprovision, deprovision. Toggle drift detected to be notified when drift is found.',
  install_configurations: 'Install input updates and secret syncs.',
  runners: 'Runner provision and reprovision.',
  actions: 'Action workflow runs.',
  app_branches: 'App branch runs triggered manually or by VCS updates.',
}

export const OUTCOME_LABELS: Record<Outcome, string> = {
  none: 'Off',
  all: 'All activity',
  completion: 'On completion',
  failures: 'On failures',
}
