export default {
  title: 'Approvals/PlanDiffs/AppConfigDiff',
}

import { AppConfigDiff } from './AppConfigDiff'
import type { DiffSectionData } from './AppConfigDiff'

const mockSections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    grouped: true,
    additions: 1,
    removals: 1,
    changed: 1,
    entities: [
      {
        name: 'redis',
        op: 'add',
        componentType: 'helm_chart',
        fields: [
          { key: 'type', op: 'add', diff: "'' -> 'helm_chart'" },
          { key: 'chart_name', op: 'add', diff: "'' -> 'redis'" },
          { key: 'namespace', op: 'add', diff: "'' -> 'cache'" },
        ],
      },
      {
        name: 'legacy-worker',
        op: 'remove',
        componentType: 'docker_build',
        fields: [
          { key: 'type', op: 'remove', diff: "'docker_build' -> ''" },
          { key: 'dockerfile', op: 'remove', diff: "'Dockerfile.worker' -> ''" },
        ],
      },
      {
        name: 'ctl-api',
        op: 'change',
        componentType: 'helm_chart',
        fields: [
          { key: 'chart_name', op: 'change', diff: "'ctl-api-v1' -> 'ctl-api-v2'" },
          { key: 'namespace', op: 'change', diff: "'default' -> 'app'" },
        ],
      },
    ],
    fields: [],
  },
  {
    name: 'Actions',
    sectionKey: 'actions',
    grouped: true,
    additions: 1,
    removals: 0,
    changed: 1,
    entities: [
      {
        name: 'run-migrations',
        op: 'add',
        fields: [
          { key: 'timeout', op: 'add', diff: "'' -> '300s'" },
          { key: 'role', op: 'add', diff: "'' -> 'admin'" },
        ],
      },
      {
        name: 'healthcheck',
        op: 'change',
        fields: [
          { key: 'role', op: 'change', diff: "'admin' -> 'operator'" },
        ],
      },
    ],
    fields: [],
  },
  {
    name: 'Runner',
    sectionKey: 'runner',
    grouped: false,
    additions: 0,
    removals: 0,
    changed: 2,
    entities: [],
    fields: [
      { key: 'runner_type', op: 'change', diff: "'standard' -> 'gpu'" },
      { key: 'init_script', op: 'change', diff: "'setup.sh' -> 'setup-gpu.sh'" },
    ],
  },
]

export const Default = () => (
  <AppConfigDiff
    sections={mockSections}
    summary={{ added: 2, removed: 1, changed: 3 }}
  />
)

export const NoChanges = () => (
  <AppConfigDiff sections={[]} summary={null} />
)

export const Loading = () => (
  <AppConfigDiff sections={[]} summary={null} isLoading />
)

export const ComponentsOnly = () => (
  <AppConfigDiff
    sections={[mockSections[0]]}
    summary={{ added: 1, removed: 1, changed: 1 }}
  />
)

export const FieldsOnly = () => (
  <AppConfigDiff
    sections={[mockSections[2]]}
    summary={{ added: 0, removed: 0, changed: 2 }}
  />
)

export const AllSections = () => {
  const allSections: DiffSectionData[] = [
    mockSections[0],
    mockSections[1],
    {
      name: 'Install inputs',
      sectionKey: 'inputs',
      grouped: true,
      additions: 2,
      removals: 0,
      changed: 0,
      entities: [
        {
          name: 'cluster_name',
          op: 'add',
          fields: [
            { key: 'type', op: 'add', diff: "'' -> 'string'" },
            { key: 'required', op: 'add', diff: "'false' -> 'true'" },
          ],
        },
        {
          name: 'region',
          op: 'add',
          fields: [
            { key: 'default', op: 'add', diff: "'' -> 'us-west-2'" },
          ],
        },
      ],
      fields: [],
    },
    {
      name: 'Secrets',
      sectionKey: 'secrets',
      grouped: true,
      additions: 1,
      removals: 0,
      changed: 0,
      entities: [
        {
          name: 'DATABASE_URL',
          op: 'add',
          fields: [
            { key: 'required', op: 'add', diff: "'' -> 'true'" },
          ],
        },
      ],
      fields: [],
    },
    mockSections[2],
    {
      name: 'Stack',
      sectionKey: 'stack',
      grouped: false,
      additions: 0,
      removals: 0,
      changed: 1,
      entities: [],
      fields: [
        { key: 'type', op: 'change', diff: "'eks' -> 'eks-v2'" },
      ],
    },
    {
      name: 'Sandbox',
      sectionKey: 'sandbox',
      grouped: false,
      additions: 0,
      removals: 0,
      changed: 1,
      entities: [],
      fields: [
        { key: 'terraform_version', op: 'change', diff: "'1.5.0' -> '1.6.0'" },
      ],
    },
    {
      name: 'Permissions',
      sectionKey: 'permissions',
      grouped: false,
      additions: 1,
      removals: 0,
      changed: 0,
      entities: [],
      fields: [
        { key: 'provision', op: 'add', diff: "'' -> 'arn:aws:iam::role/deploy'" },
      ],
    },
  ]

  return (
    <AppConfigDiff
      sections={allSections}
      summary={{ added: 6, removed: 1, changed: 5 }}
    />
  )
}
