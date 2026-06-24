export default {
  title: 'Approvals/PlanDiffs/AppConfigDiff',
}

import { AppConfigDiff } from './AppConfigDiff'
import type { DiffSectionData } from './AppConfigDiff'

const mockSections: DiffSectionData[] = [
  {
    name: 'Components',
    additions: 1,
    removals: 1,
    changed: 2,
    entries: [
      { op: 'add', name: 'redis', description: 'helm_chart' },
      { op: 'remove', name: 'legacy-worker', description: 'docker_build' },
      { op: 'change', name: 'ctl-api', description: "chart_name: 'ctl-api-v1' -> 'ctl-api-v2'" },
      { op: 'change', name: 'dashboard', description: "namespace: 'default' -> 'app'" },
    ],
  },
  {
    name: 'Actions',
    additions: 1,
    removals: 0,
    changed: 1,
    entries: [
      { op: 'add', name: 'run-migrations', description: 'timeout: 300s' },
      { op: 'change', name: 'healthcheck', description: "role: 'admin' -> 'operator'" },
    ],
  },
  {
    name: 'Install inputs',
    additions: 2,
    removals: 0,
    changed: 0,
    entries: [
      { op: 'add', name: 'cluster_name', description: 'required string input' },
      { op: 'add', name: 'region', description: "default: 'us-west-2'" },
    ],
  },
]

export const Default = () => (
  <AppConfigDiff
    sections={mockSections}
    summary={{ added: 4, removed: 1, changed: 3 }}
  />
)

export const NoChanges = () => (
  <AppConfigDiff
    sections={[]}
    summary={null}
  />
)

export const Loading = () => (
  <AppConfigDiff
    sections={[]}
    summary={null}
    isLoading
  />
)

export const SingleSection = () => (
  <AppConfigDiff
    sections={[mockSections[0]]}
    summary={{ added: 1, removed: 1, changed: 2 }}
  />
)

export const ManyChanges = () => {
  const largeSections: DiffSectionData[] = [
    {
      name: 'Components',
      additions: 5,
      removals: 2,
      changed: 8,
      entries: [
        { op: 'add', name: 'redis', description: 'helm_chart' },
        { op: 'add', name: 'postgres', description: 'helm_chart' },
        { op: 'add', name: 'monitoring', description: 'terraform_module' },
        { op: 'add', name: 'cert-manager', description: 'helm_chart' },
        { op: 'add', name: 'ingress-nginx', description: 'helm_chart' },
        { op: 'remove', name: 'legacy-worker', description: 'docker_build' },
        { op: 'remove', name: 'old-proxy', description: 'docker_build' },
        { op: 'change', name: 'ctl-api', description: "chart_name: 'ctl-api-v1' -> 'ctl-api-v2'" },
        { op: 'change', name: 'dashboard', description: "namespace: 'default' -> 'app'" },
        { op: 'change', name: 'worker', description: "replicas: '1' -> '3'" },
        { op: 'change', name: 'api-gateway', description: "image_tag: 'v1.2' -> 'v2.0'" },
        { op: 'change', name: 'scheduler', description: "cron: '*/5 * * * *' -> '*/2 * * * *'" },
        { op: 'change', name: 'notifier', description: "env.SLACK_CHANNEL: '#alerts' -> '#ops-alerts'" },
        { op: 'change', name: 'auth-service', description: "timeout: '30s' -> '60s'" },
        { op: 'change', name: 'cache', description: "max_memory: '256Mi' -> '512Mi'" },
      ],
    },
    {
      name: 'Secrets',
      additions: 3,
      removals: 0,
      changed: 0,
      entries: [
        { op: 'add', name: 'DATABASE_URL', description: 'required, no default' },
        { op: 'add', name: 'REDIS_PASSWORD', description: 'required, auto_generate' },
        { op: 'add', name: 'API_KEY', description: 'required, format: uuid' },
      ],
    },
    {
      name: 'Sandbox',
      additions: 0,
      removals: 0,
      changed: 2,
      entries: [
        { op: 'change', name: 'terraform_version', description: "'1.5.0' -> '1.6.0'" },
        { op: 'change', name: 'drift_schedule', description: "'@daily' -> '@hourly'" },
      ],
    },
  ]

  return (
    <AppConfigDiff
      sections={largeSections}
      summary={{ added: 8, removed: 2, changed: 10 }}
    />
  )
}
