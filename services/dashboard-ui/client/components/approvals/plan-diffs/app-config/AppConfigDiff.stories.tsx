export default {
  title: 'Approvals/PlanDiffs/AppConfigDiff',
}

import { AppConfigDiff } from './AppConfigDiff'
import type { DiffSectionData } from './AppConfigDiff'

const mockSections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    additions: 1,
    removals: 1,
    changed: 2,
    entries: [
      { op: 'add', name: 'redis', description: 'helm_chart' },
      { op: 'remove', name: 'legacy-worker', description: 'docker_build' },
      { op: 'change', name: 'ctl-api', description: 'terraform_module' },
      { op: 'change', name: 'dashboard', description: 'kubernetes_manifest' },
    ],
  },
  {
    name: 'Actions',
    sectionKey: 'actions',
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
    sectionKey: 'inputs',
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

export const AllSections = () => {
  const allSections: DiffSectionData[] = [
    {
      name: 'Components',
      sectionKey: 'components',
      additions: 2,
      removals: 1,
      changed: 3,
      entries: [
        { op: 'add', name: 'redis', description: 'helm_chart' },
        { op: 'add', name: 'monitoring', description: 'terraform_module' },
        { op: 'remove', name: 'legacy-worker', description: 'docker_build' },
        { op: 'change', name: 'ctl-api', description: 'helm_chart' },
        { op: 'change', name: 'runner', description: 'external_image' },
        { op: 'change', name: 'infra', description: 'pulumi' },
      ],
    },
    {
      name: 'Actions',
      sectionKey: 'actions',
      additions: 1,
      removals: 0,
      changed: 0,
      entries: [
        { op: 'add', name: 'run-migrations', description: 'timeout: 300s' },
      ],
    },
    {
      name: 'Install inputs',
      sectionKey: 'inputs',
      additions: 2,
      removals: 0,
      changed: 0,
      entries: [
        { op: 'add', name: 'cluster_name', description: 'required string' },
        { op: 'add', name: 'region', description: "default: 'us-west-2'" },
      ],
    },
    {
      name: 'Secrets',
      sectionKey: 'secrets',
      additions: 1,
      removals: 0,
      changed: 0,
      entries: [
        { op: 'add', name: 'DATABASE_URL', description: 'required, no default' },
      ],
    },
    {
      name: 'Sandbox',
      sectionKey: 'sandbox',
      additions: 0,
      removals: 0,
      changed: 1,
      entries: [
        { op: 'change', name: 'terraform_version', description: "'1.5.0' -> '1.6.0'" },
      ],
    },
    {
      name: 'Runner',
      sectionKey: 'runner',
      additions: 0,
      removals: 0,
      changed: 1,
      entries: [
        { op: 'change', name: 'runner_type', description: "'standard' -> 'gpu'" },
      ],
    },
    {
      name: 'Permissions',
      sectionKey: 'permissions',
      additions: 1,
      removals: 0,
      changed: 0,
      entries: [
        { op: 'add', name: 'provision', description: 'arn:aws:iam::role/deploy' },
      ],
    },
    {
      name: 'Stack',
      sectionKey: 'stack',
      additions: 0,
      removals: 0,
      changed: 1,
      entries: [
        { op: 'change', name: 'type', description: "'eks' -> 'eks-v2'" },
      ],
    },
  ]

  return (
    <AppConfigDiff
      sections={allSections}
      summary={{ added: 7, removed: 1, changed: 6 }}
    />
  )
}
