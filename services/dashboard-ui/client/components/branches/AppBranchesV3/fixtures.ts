import type { TRunSource } from '@/components/branches/BranchOverview/run-source'
import type {
  TAppBranchInstallGroup,
  TAppConfigDiffSection,
  TCompositeError,
} from '@/types'
import type { TConfigSourceFile } from './ConfigChanges/config-changes'

export type TRolloutInstall = {
  id: string
  name: string
  status: string
  configVersion: number
  targetConfigVersion: number
  awaitingApproval?: boolean
  compositeError?: TCompositeError
}

const cacheDeployError: TCompositeError = {
  version: 1,
  type: 'deploy.apply_failed',
  severity: 'error',
  message: 'Deploying cache failed',
  sections: [
    {
      heading: 'Why',
      kind: 'markdown',
      body: 'The cache release failed while creating its service. The api and worker releases were not updated.',
    },
    {
      heading: 'Error detail',
      kind: 'code',
      body: 'Error: Service "cache" is invalid: spec.ports[0].port: Invalid value: 0: must be between 1 and 65535',
    },
  ],
}

const install = (
  id: string,
  name: string,
  status: string,
  awaitingApproval = false
): TRolloutInstall => ({
  id,
  name,
  status,
  configVersion: status === 'success' ? 14 : 13,
  targetConfigVersion: 14,
  awaitingApproval,
})

export type TRolloutInstallGroup = TAppBranchInstallGroup & {
  status: string
  installs: TRolloutInstall[]
}

export type TConfigChangeSummary = {
  added: number
  removed: number
  changed: number
}

export type TLatestRollout = {
  id: string
  status: string
  title: string
  source: TRunSource
  configChanges: {
    versionLabel: string
    summary: TConfigChangeSummary
    sections: TAppConfigDiffSection[]
    files: TConfigSourceFile[]
  }
  installGroups: TRolloutInstallGroup[]
  commit: {
    message: string
    author: string
    sha: string
    previousSha: string
    shaUrl: string
    createdAt: string
  }
}

const MINUTE = 60_000

const ago = (minutes: number) =>
  new Date(Date.now() - minutes * MINUTE).toISOString()

export const latestRolloutFixture: TLatestRollout = {
  id: 'run-latest',
  status: 'in-progress',
  installGroups: [
    {
      id: 'grp-canary',
      name: 'Canary',
      order: 1,
      max_parallel: 2,
      auto_approve_on_policies_passing: true,
      label_selector: { match_labels: { env: 'prod', tier: 'canary' } },
      status: 'success',
      installs: [
        install('inst-alpha', 'alpha', 'success'),
        install('inst-bravo', 'bravo', 'success'),
      ],
    },
    {
      id: 'grp-primary',
      name: 'Primary region',
      order: 2,
      max_parallel: 2,
      auto_approve_on_policies_passing: false,
      label_selector: { match_labels: { env: 'prod', region: 'us-east-1' } },
      status: 'in-progress',
      installs: [
        install('inst-charlie', 'charlie', 'success'),
        install('inst-hotel', 'hotel', 'success'),
        install('inst-india', 'india', 'success'),
        install('inst-juliet', 'juliet', 'success'),
        {
          ...install('inst-delta', 'delta', 'error'),
          compositeError: cacheDeployError,
        },
        install('inst-echo', 'echo', 'in-progress', true),
        install('inst-foxtrot', 'foxtrot', 'in-progress', true),
      ],
    },
    {
      id: 'grp-eu-west',
      name: 'EU west',
      order: 3,
      max_parallel: 3,
      auto_approve_on_policies_passing: true,
      label_selector: { match_labels: { env: 'prod', region: 'eu-west-1' } },
      status: 'in-progress',
      installs: [
        install('inst-kilo', 'kilo', 'success'),
        install('inst-lima', 'lima', 'success'),
        install('inst-mike', 'mike', 'success'),
        install('inst-november', 'november', 'success'),
        install('inst-oscar', 'oscar', 'success'),
        install('inst-papa', 'papa', 'in-progress', true),
      ],
    },
    {
      id: 'grp-staging',
      name: 'Staging',
      order: 4,
      max_parallel: 2,
      auto_approve_on_policies_passing: true,
      label_selector: { match_labels: { env: 'staging' } },
      status: 'in-progress',
      installs: [
        install('inst-quebec', 'quebec', 'success'),
        {
          ...install('inst-romeo', 'romeo', 'error'),
          compositeError: cacheDeployError,
        },
        install('inst-sierra', 'sierra', 'in-progress'),
        install('inst-tango', 'tango', 'in-progress'),
      ],
    },
    {
      id: 'grp-dedicated',
      name: 'Dedicated enterprise tenants in regulated regions',
      order: 5,
      max_parallel: 1,
      auto_approve_on_policies_passing: false,
      label_selector: {
        match_labels: {
          tenancy: 'dedicated',
          compliance_profile: 'regulated-financial-services-tier-2',
          deployment_ring: 'enterprise-early-access',
          workspace_id: 'workspace_7k2m9qx4vb8np3rt6yw1hs5dce',
          customer_account_id: 'acct_4f8e2a9c1b7d3e6f0a5b8c2d9e1f',
          maintenance_window: 'sunday-0200-0600-utc',
          data_residency: 'eu-central',
        },
      },
      status: 'pending',
      installs: [
        install(
          'inst-tenant-1',
          'acme-dedicated-eu-central-1-prod-tenant-northwind-financial',
          'pending'
        ),
        install(
          'inst-tenant-2',
          'workspace_7k2m9qx4vb8np3rt6yw1hs5dce',
          'pending'
        ),
        install(
          'inst-tenant-3',
          'acme-dedicated-eu-west-2-prod-tenant-contoso-holdings',
          'pending'
        ),
      ],
    },
    {
      id: 'grp-default',
      name: 'Remaining',
      order: 6,
      max_parallel: 4,
      default: true,
      auto_approve_on_policies_passing: true,
      status: 'in-progress',
      installs: [
        install('inst-rest-1', 'install-1', 'success'),
        install('inst-rest-2', 'install-2', 'success'),
        install('inst-rest-3', 'install-3', 'in-progress'),
        install('inst-rest-4', 'install-4', 'in-progress'),
        ...Array.from({ length: 10 }, (_, index) =>
          install(`inst-rest-${index + 5}`, `install-${index + 5}`, 'pending')
        ),
      ],
    },
  ],
  title: 'Add cache component',
  source: {
    kind: 'pull-request',
    number: 482,
    url: 'https://github.com/acme/platform/pull/482',
    baseBranch: 'main',
  },
  commit: {
    message:
      'Add cache component\n\nAdds a shared cache module and wires the api and worker charts to it.',
    author: 'jane@example.com',
    sha: 'a1b2c3d4e5f60718293a4b5c6d7e8f9012345678',
    previousSha: '9f8e7d6c5b4a39281706f5e4d3c2b1a098765432',
    shaUrl:
      'https://github.com/acme/platform/commit/a1b2c3d4e5f60718293a4b5c6d7e8f9012345678',
    createdAt: ago(12),
  },
  configChanges: {
    versionLabel: 'v13 → v14',
    summary: { added: 2, removed: 1, changed: 3 },
    sections: [
      {
        name: 'Components',
        sectionKey: 'components',
        additions: 1,
        removals: 1,
        changed: 2,
        grouped: true,
        fields: [],
        entities: [
          {
            name: 'cache',
            op: 'add',
            componentType: 'helm_chart',
            fields: [
              { key: 'type', op: 'add', diff: "'helm_chart'" },
              { key: 'chart_name', op: 'add', diff: "'cache'" },
              { key: 'namespace', op: 'add', diff: "'acme'" },
              {
                key: 'public_repo.repo',
                op: 'add',
                diff: "'acme/platform'",
              },
              {
                key: 'public_repo.directory',
                op: 'add',
                diff: "'charts/cache'",
              },
            ],
          },
          {
            name: 'api',
            op: 'change',
            componentType: 'helm_chart',
            fields: [
              {
                key: 'public_repo.branch',
                op: 'change',
                diff: "'release-13' -> 'release-14'",
              },
              {
                key: 'dependencies',
                op: 'change',
                diff: "'legacy-redis' -> 'cache'",
              },
            ],
          },
          {
            name: 'worker',
            op: 'change',
            componentType: 'helm_chart',
            fields: [
              {
                key: 'values.cache_endpoint',
                op: 'change',
                diff: "'{{ .nuon.components.legacy-redis.outputs.host }}' -> '{{ .nuon.components.cache.outputs.endpoint }}'",
              },
              { key: 'values.replicas', op: 'change', diff: "'2' -> '3'" },
            ],
          },
          {
            name: 'legacy-redis',
            op: 'remove',
            componentType: 'helm_chart',
            fields: [
              { key: 'type', op: 'remove', diff: "'helm_chart'" },
              { key: 'chart_name', op: 'remove', diff: "'redis'" },
              { key: 'namespace', op: 'remove', diff: "'acme'" },
            ],
          },
        ],
      },
      {
        name: 'Inputs',
        sectionKey: 'inputs',
        additions: 1,
        removals: 0,
        changed: 0,
        grouped: true,
        fields: [],
        entities: [
          {
            name: 'cache_size',
            op: 'add',
            fields: [
              { key: 'display_name', op: 'add', diff: "'Cache size'" },
              { key: 'default', op: 'add', diff: "'2Gi'" },
              { key: 'group', op: 'add', diff: "'cache'" },
            ],
          },
        ],
      },
      {
        name: 'Sandbox',
        sectionKey: 'sandbox',
        additions: 0,
        removals: 0,
        changed: 1,
        grouped: false,
        entities: [],
        fields: [],
        content: {
          op: 'change',
          before:
            'terraform_version = "1.9.5"\n\n[public_repo]\nrepo = "acme/sandboxes"\ndirectory = "eks"\nbranch = "v3.2.0"\n\n[vars]\ncluster_version = "1.30"\nnode_count = 3\n',
          after:
            'terraform_version = "1.9.5"\n\n[public_repo]\nrepo = "acme/sandboxes"\ndirectory = "eks"\nbranch = "v3.3.0"\n\n[vars]\ncluster_version = "1.31"\nnode_count = 3\n',
        },
      },
    ],
    files: [
      {
        path: 'values/cache.yaml',
        kind: 'helm values',
        change: 'added',
        after:
          'replicaCount: 2\n\nresources:\n  requests:\n    memory: "{{ .nuon.inputs.inputs.cache_size }}"\n\nservice:\n  port: 6379\n',
      },
      {
        path: 'values/api.yaml',
        kind: 'helm values',
        change: 'modified',
        before:
          'image:\n  repository: acme/api\n  tag: "1.4.1"\n\nenv:\n  REDIS_HOST: "{{ .nuon.components.legacy-redis.outputs.host }}"\n  LOG_LEVEL: info\n',
        after:
          'image:\n  repository: acme/api\n  tag: "1.4.2"\n\nenv:\n  CACHE_ENDPOINT: "{{ .nuon.components.cache.outputs.endpoint }}"\n  LOG_LEVEL: info\n',
      },
      {
        path: 'values/legacy-redis.yaml',
        kind: 'helm values',
        change: 'removed',
        before: 'architecture: standalone\n\nauth:\n  enabled: false\n',
      },
    ],
  },
}

type TInstallOutcome = (
  install: TRolloutInstall,
  group: TRolloutInstallGroup
) => string

const finishedGroups = (
  version: number,
  outcome: TInstallOutcome
): TRolloutInstallGroup[] =>
  latestRolloutFixture.installGroups.map((group) => {
    const installs = group.installs.map((install) => {
      const status = outcome(install, group)
      return {
        id: install.id,
        name: install.name,
        status,
        configVersion: status === 'success' ? version : version - 1,
        targetConfigVersion: version,
      }
    })
    const statuses = new Set(installs.map(({ status }) => status))
    return {
      ...group,
      installs,
      status: statuses.has('error')
        ? 'error'
        : statuses.has('success')
          ? 'success'
          : 'cancelled',
    }
  })

type TOlderRollout = {
  id: string
  version: number
  sha: string
  previousSha: string
  message: string
  author: string
  minutesAgo: number
  source: TRunSource
  status: string
  outcome: TInstallOutcome
  summary: TConfigChangeSummary
}

const olderRollout = ({
  id,
  version,
  sha,
  previousSha,
  message,
  author,
  minutesAgo,
  source,
  status,
  outcome,
  summary,
}: TOlderRollout): TLatestRollout => ({
  ...latestRolloutFixture,
  id,
  status,
  title: message.split('\n')[0],
  source,
  commit: {
    message,
    author,
    sha,
    previousSha,
    shaUrl: `https://github.com/acme/platform/commit/${sha}`,
    createdAt: ago(minutesAgo),
  },
  configChanges: {
    ...latestRolloutFixture.configChanges,
    versionLabel: `v${version - 1} → v${version}`,
    summary,
  },
  installGroups: finishedGroups(version, outcome),
})

const HOUR = 60
const DAY = 24 * HOUR

const NEVER_APPROVED = new Set(['inst-echo', 'inst-foxtrot', 'inst-papa'])

export const previousRolloutFixtures: TLatestRollout[] = [
  olderRollout({
    id: 'run-13',
    version: 13,
    sha: '9f8e7d6c5b4a39281706f5e4d3c2b1a098765432',
    previousSha: '5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f',
    message:
      'Bump worker replicas\n\nRaises the default worker replica count for the queue backlog.',
    author: 'sam@example.com',
    minutesAgo: 3 * HOUR,
    source: {
      kind: 'pull-request',
      number: 478,
      url: 'https://github.com/acme/platform/pull/478',
      baseBranch: 'main',
    },
    status: 'success',
    outcome: (install) =>
      NEVER_APPROVED.has(install.id) ? 'cancelled' : 'success',
    summary: { added: 0, removed: 0, changed: 1 },
  }),
  olderRollout({
    id: 'run-12',
    version: 12,
    sha: '5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f',
    previousSha: '1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d',
    message:
      'Move sandbox to eks module v3.2.0\n\nPicks up the node group autoscaling fix.',
    author: 'jane@example.com',
    minutesAgo: 2 * DAY,
    source: { kind: 'tag', tag: 'v2.8.0' },
    status: 'error',
    outcome: (install, group) => {
      if (install.id === 'inst-delta' || install.id === 'inst-romeo') {
        return 'error'
      }
      if (group.id === 'grp-dedicated') return 'cancelled'
      return 'success'
    },
    summary: { added: 0, removed: 0, changed: 2 },
  }),
  olderRollout({
    id: 'run-11',
    version: 11,
    sha: '1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d',
    previousSha: '0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c',
    message: 'Add region input',
    author: 'sam@example.com',
    minutesAgo: 5 * DAY,
    source: { kind: 'manual' },
    status: 'success',
    outcome: () => 'success',
    summary: { added: 1, removed: 0, changed: 0 },
  }),
  olderRollout({
    id: 'run-10',
    version: 10,
    sha: '0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c',
    previousSha: '7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d',
    message:
      'Remove legacy cron action\n\nThe nightly cleanup now runs inside the worker.',
    author: 'alex@example.com',
    minutesAgo: 9 * DAY,
    source: { kind: 'commit' },
    status: 'success',
    outcome: (install, group) =>
      group.id === 'grp-primary' && NEVER_APPROVED.has(install.id)
        ? 'cancelled'
        : 'success',
    summary: { added: 0, removed: 1, changed: 0 },
  }),
]
