import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import type { TRunSource } from '@/components/branches/BranchOverview/run-source'
import type { TAppBranchInstallGroup, TCompositeError } from '@/types'

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
  status: string
  title: string
  source: TRunSource
  configChanges: {
    versionLabel: string
    summary: TConfigChangeSummary
    sections: DiffSectionData[]
  }
  installGroups: TRolloutInstallGroup[]
  commit: {
    message: string
    author: string
    sha: string
    shaUrl: string
    createdAt: string
  }
}

const MINUTE = 60_000

const ago = (minutes: number) =>
  new Date(Date.now() - minutes * MINUTE).toISOString()

export const latestRolloutFixture: TLatestRollout = {
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
      id: 'grp-dedicated',
      name: 'Dedicated enterprise tenants in regulated regions',
      order: 3,
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
      order: 4,
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
    shaUrl:
      'https://github.com/acme/platform/commit/a1b2c3d4e5f60718293a4b5c6d7e8f9012345678',
    createdAt: ago(12),
  },
  configChanges: {
    versionLabel: 'v13 → v14',
    summary: { added: 1, removed: 0, changed: 2 },
    sections: [
      {
        name: 'Components',
        sectionKey: 'components',
        additions: 1,
        removals: 0,
        changed: 2,
        grouped: true,
        fields: [],
        entities: [
          {
            name: 'cache',
            op: 'add',
            componentType: 'terraform_module',
            fields: [
              {
                key: 'terraform.version',
                op: 'add',
                diff: "'' -> '1.9.5'",
              },
            ],
          },
          {
            name: 'api',
            op: 'change',
            componentType: 'docker_build',
            fields: [
              {
                key: 'image.tag',
                op: 'change',
                diff: "'1.4.1' -> '1.4.2'",
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
                diff: "'' -> '{{ .nuon.components.cache.outputs.endpoint }}'",
              },
            ],
          },
        ],
      },
    ],
  },
}
