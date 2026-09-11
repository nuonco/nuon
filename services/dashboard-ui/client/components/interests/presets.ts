import { allEvents } from './defaults'
import type { Interests } from './types'
import type { TargetKind } from '@/components/match/types'

export type PresetId = 'all' | 'failures' | 'operations' | 'approvals' | 'deploys' | 'custom'

export interface Preset {
  id: PresetId
  label: string
  description: string
  build: () => Interests
  recommended?: boolean
  scopeKind?: TargetKind
}

const buildFailuresOnly = (): Interests => ({
  resources: {
    installs: { outcome: 'failures' },
    stacks: { outcome: 'failures' },
    components: { outcome: 'failures' },
    sandboxes: { outcome: 'failures' },
    install_configurations: { outcome: 'failures' },
    runners: { outcome: 'failures' },
    actions: { outcome: 'failures' },
    app_branches: { outcome: 'failures' },
  },
})

const buildOperations = (): Interests => ({
  resources: {
    installs: { outcome: 'failures' },
    stacks: { outcome: 'failures' },
    components: {
      outcome: 'failures',
      drift_detected: true,
    },
    sandboxes: {
      outcome: 'failures',
      drift_detected: true,
    },
    install_configurations: { outcome: 'failures' },
    runners: { outcome: 'failures' },
    actions: { outcome: 'failures' },
    app_branches: { outcome: 'failures' },
  },
})

const buildApprovalsOnly = (): Interests => ({
  resources: {
    installs: {
      outcome: 'none',
      approval_requests: true,
      approval_responses: true,
    },
    components: {
      outcome: 'none',
      approval_requests: true,
      approval_responses: true,
    },
    sandboxes: {
      outcome: 'none',
      approval_requests: true,
      approval_responses: true,
    },
    install_configurations: {
      outcome: 'none',
      approval_requests: true,
      approval_responses: true,
    },
  },
})

const buildDeploymentMilestones = (): Interests => ({
  resources: {
    installs: { outcome: 'completion' },
  },
})

export const PRESETS: Preset[] = [
  {
    id: 'failures',
    label: 'Failures',
    description:
      'Only notify when lifecycle events fail or are cancelled.',
    build: buildFailuresOnly,
    recommended: true,
    scopeKind: 'installs',
  },
  {
    id: 'all',
    label: 'All events',
    description: 'Get notified about every event — lifecycle, approvals, drift, and config syncs.',
    build: allEvents,
  },
  {
    id: 'operations',
    label: 'Operations',
    description:
      'Failed lifecycle events and drift detected — for the ops channel that needs to act on problems.',
    build: buildOperations,
  },
  {
    id: 'approvals',
    label: 'Approvals only',
    description: 'Only notify when an approval is requested or resolved across all resources.',
    build: buildApprovalsOnly,
  },
  {
    id: 'deploys',
    label: 'Deployment milestones',
    description:
      'Notify when a customer install is provisioned, deprovisioned, or reprovisioned.',
    build: buildDeploymentMilestones,
    scopeKind: 'installs',
  },
  {
    id: 'custom',
    label: 'Custom',
    description: 'Pick exactly which events you want for each resource type.',
    build: () => ({}),
  },
]

const canonicalize = (v: unknown): unknown => {
  if (v === null || typeof v !== 'object') return v
  if (Array.isArray(v)) return v.map(canonicalize)
  const sorted: Record<string, unknown> = {}
  for (const key of Object.keys(v as Record<string, unknown>).sort()) {
    sorted[key] = canonicalize((v as Record<string, unknown>)[key])
  }
  return sorted
}

const deepEqual = (a: unknown, b: unknown): boolean =>
  JSON.stringify(canonicalize(a)) === JSON.stringify(canonicalize(b))

export const recommendedPreset = (): Preset =>
  PRESETS.find((p) => p.recommended) ?? PRESETS[0]

export const matchPreset = (value: Interests): PresetId => {
  if (value.all_events) return 'all'
  for (const preset of PRESETS) {
    if (preset.id === 'all' || preset.id === 'custom') continue
    if (deepEqual(value, preset.build())) return preset.id
  }
  return 'custom'
}
