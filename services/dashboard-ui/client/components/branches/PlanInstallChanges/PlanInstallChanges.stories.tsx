export default {
  title: 'Features / Branches / Plan Install Changes',
}

import type { ReactNode } from 'react'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { PlanInstallDiff } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/PlanGroupStep'
import {
  PlanInstallChanges,
  type TPlanInstallFacts,
} from './PlanInstallChanges'

const NEXT = 'c4d5e6f708192a3b'
const PREVIOUS = '9f8e7d6c5b4a3210'
const OLDER = '1122334455667788'

const components: DiffSectionData = {
  name: 'Components',
  sectionKey: 'components',
  additions: 1,
  removals: 1,
  changed: 1,
  grouped: true,
  fields: [],
  entities: [
    {
      name: 'cache',
      op: 'add',
      componentType: 'helm_chart',
      fields: [
        { key: 'chart_name', op: 'add', diff: "'cache'" },
        { key: 'namespace', op: 'add', diff: "'acme'" },
        { key: 'values.replicaCount', op: 'add', diff: "'2'" },
      ],
    },
    {
      name: 'api',
      op: 'change',
      componentType: 'helm_chart',
      fields: [
        { key: 'values.replicaCount', op: 'change', diff: "'1' -> '3'" },
        {
          key: 'public_repo.tag',
          op: 'change',
          diff: "'release-13' -> 'release-14'",
        },
      ],
    },
    {
      name: 'legacy-redis',
      op: 'remove',
      componentType: 'helm_chart',
      fields: [{ key: 'chart_name', op: 'remove', diff: "'redis'" }],
    },
  ],
}

const runnerAndInputs: DiffSectionData[] = [
  {
    name: 'Runner',
    sectionKey: 'runner',
    additions: 0,
    removals: 0,
    changed: 1,
    grouped: false,
    entities: [],
    fields: [
      { key: 'image.tag', op: 'change', diff: "'1.8.2' -> '1.9.0'" },
      { key: 'cpu', op: 'change', diff: "'500m' -> '1000m'" },
    ],
  },
  {
    name: 'Install inputs',
    sectionKey: 'inputs',
    additions: 1,
    removals: 0,
    changed: 1,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'log_level',
        op: 'change',
        fields: [{ key: 'value', op: 'change', diff: "'info' -> 'warn'" }],
      },
      {
        name: 'cache_enabled',
        op: 'add',
        fields: [{ key: 'value', op: 'add', diff: "'true'" }],
      },
    ],
  },
]

const sandbox: DiffSectionData = {
  name: 'Sandbox',
  sectionKey: 'sandbox',
  additions: 0,
  removals: 0,
  changed: 1,
  grouped: false,
  entities: [],
  fields: [
    {
      key: 'terraform.version',
      op: 'change',
      diff: "'1.9.0' -> '1.10.2'",
    },
    {
      key: 'network.vpc_cidr',
      op: 'change',
      diff: "'10.0.0.0/16' -> '10.1.0.0/16'",
    },
  ],
}

const permissions: DiffSectionData = {
  name: 'Permissions',
  sectionKey: 'permissions',
  additions: 1,
  removals: 0,
  changed: 0,
  grouped: false,
  entities: [],
  fields: [
    {
      key: 'iam.actions',
      op: 'add',
      diff: "'elasticache:DescribeCacheClusters'",
    },
  ],
}

const releaseCommit = {
  sha: NEXT,
  previousSha: PREVIOUS,
  message: 'Raise api replicas and add the cache chart',
  author: 'Example Developer',
  createdAt: '2026-03-12T15:04:00Z',
}

const install = (
  id: string,
  name: string,
  sections: DiffSectionData[],
  summary: PlanInstallDiff['summary'],
  versionLabel: string,
  commit: Partial<
    Pick<PlanInstallDiff, 'sha' | 'previousSha' | 'message' | 'author' | 'createdAt'>
  > = releaseCommit,
  labels?: Record<string, string>
): PlanInstallDiff => ({
  installId: id,
  installName: name,
  installLabels: labels,
  sections,
  summary,
  versionLabel,
  ...commit,
})

const mixedInstalls: PlanInstallDiff[] = [
  install(
    'inst-acme-east',
    'acme-us-east',
    [components, ...runnerAndInputs, permissions],
    { added: 3, removed: 1, changed: 4 },
    'v11 → v14',
    releaseCommit,
    { env: 'prod', tier: 'enterprise' }
  ),
  install(
    'inst-northwind',
    'northwind-eu-west',
    runnerAndInputs,
    { added: 1, removed: 0, changed: 2 },
    'v13 → v14',
    { ...releaseCommit, previousSha: OLDER },
    { env: 'prod', tier: 'standard' }
  ),
  install(
    'inst-globex',
    'globex-ap-south',
    [sandbox, permissions],
    { added: 1, removed: 0, changed: 1 },
    'v8 → v14',
    {
      sha: NEXT,
      previousSha: 'aabbccddeeff0011',
      message: 'Bump the sandbox Terraform version',
      author: 'Example Developer',
      createdAt: '2026-02-02T09:30:00Z',
    },
    { env: 'prod', tier: 'enterprise' }
  ),
  install(
    'inst-initech',
    'initech-us-west',
    [],
    null,
    'v14',
    { sha: NEXT, previousSha: undefined },
    { env: 'prod', tier: 'standard' }
  ),
]

const mixedFacts: TPlanInstallFacts = {
  'inst-acme-east': { region: 'us-east-1', status: 'pending' },
  'inst-northwind': { region: 'eu-west-1', status: 'pending' },
  'inst-globex': { region: 'ap-southeast-2', status: 'pending' },
  'inst-initech': { region: 'us-west-2', status: 'active' },
}

const customers = [
  'acme',
  'northwind',
  'globex',
  'initech',
  'umbrella',
  'stark',
  'wayne',
  'oscorp',
  'hooli',
  'pied-piper',
  'massive-dynamic',
  'wonka',
  'soylent',
  'tyrell',
  'cyberdyne',
  'weyland',
  'bluth',
  'dunder',
  'veridian',
  'prestigo',
]

const regions = ['us-east-1', 'us-west-2', 'eu-west-1', 'ap-southeast-2']

const largeInstalls: PlanInstallDiff[] = customers.map((customer, index) => {
  const behind = index % 5 === 0 ? 11 : index % 3 === 0 ? 13 : 14
  const unchanged = behind === 14
  return install(
    `inst-${customer}`,
    `${customer}-prod`,
    unchanged ? [] : index % 2 === 0 ? [components] : runnerAndInputs,
    unchanged
      ? null
      : index % 2 === 0
        ? { added: 1, removed: 1, changed: 1 }
        : { added: 1, removed: 0, changed: 2 },
    unchanged ? 'v14' : `v${behind} → v14`,
    unchanged
      ? { sha: NEXT }
      : { ...releaseCommit, previousSha: index % 2 === 0 ? PREVIOUS : OLDER },
    { env: 'prod' }
  )
})

const largeFacts: TPlanInstallFacts = Object.fromEntries(
  customers.map((customer, index) => [
    `inst-${customer}`,
    { region: regions[index % regions.length], status: 'pending' },
  ])
)

const Frame = ({ children }: { children: ReactNode }) => (
  <div className="flex h-[40rem] flex-col">{children}</div>
)

export const MixedGroup = () => (
  <Frame>
    <PlanInstallChanges installs={mixedInstalls} installFacts={mixedFacts} />
  </Frame>
)
MixedGroup.storyName = 'Mixed group'

export const SingleInstall = () => (
  <Frame>
    <PlanInstallChanges
      installs={mixedInstalls.slice(0, 1)}
      installFacts={mixedFacts}
    />
  </Frame>
)
SingleInstall.storyName = 'Single install'

export const LargeGroup = () => (
  <Frame>
    <PlanInstallChanges installs={largeInstalls} installFacts={largeFacts} />
  </Frame>
)
LargeGroup.storyName = 'Large group'
