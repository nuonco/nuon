// Mirrors AllEvents() and Default() in
// services/ctl-api/internal/pkg/interests/defaults.go.

import type { Interests } from './types'

// New-subscription default: matches every supported lifecycle + approval event.
export const allEvents = (): Interests => ({ all_events: true })

// Power-user opted out of AllEvents baseline. Four resources present, runners +
// actions absent. Empty ops = all sub-ops; OutcomeCompletion + both approval
// flags true. drift_detected is on for the two resources that can produce a
// drift_detected event (components, sandboxes — see RESOURCES_WITH_DRIFT_DETECTED).
export const defaultInterests = (): Interests => ({
  resources: {
    installs: { outcome: 'completion', approval_requests: true, approval_responses: true },
    stacks: { outcome: 'completion' },
    components: { outcome: 'completion', approval_requests: true, approval_responses: true, drift_detected: true },
    sandboxes: { outcome: 'completion', approval_requests: true, approval_responses: true, drift_detected: true },
    install_configurations: { outcome: 'completion', approval_requests: true, approval_responses: true },
    app_branches: { outcome: 'completion', config_synced: true },
  },
})

export const isZero = (i: Interests): boolean =>
  !i.all_events && (!i.resources || Object.keys(i.resources).length === 0)
