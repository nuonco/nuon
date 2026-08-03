// Mirrors AllEvents() and Default() in
// services/ctl-api/internal/pkg/interests/defaults.go.

import type { Interests } from './types'

// New-subscription default: matches every supported lifecycle + approval event.
export const allEvents = (): Interests => ({ all_events: true })

export const defaultInterests = (): Interests => ({
  resources: {
    installs: {
      outcome: 'completion',
      approval_requests: true,
      approval_responses: true,
      install_degraded: true,
    },
    stacks: { outcome: 'completion' },
    components: {
      outcome: 'completion',
      approval_requests: true,
      approval_responses: true,
      drift_detected: true,
      component_health: true,
    },
    sandboxes: { outcome: 'completion', approval_requests: true, approval_responses: true, drift_detected: true },
    install_configurations: { outcome: 'completion', approval_requests: true, approval_responses: true },
    app_branches: { outcome: 'completion', config_synced: true },
  },
})

export const isZero = (i: Interests): boolean =>
  !i.all_events && (!i.resources || Object.keys(i.resources).length === 0)
