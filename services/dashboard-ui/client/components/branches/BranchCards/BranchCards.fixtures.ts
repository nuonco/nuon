import type { TMiniDeployInstall } from '@/components/branches/MiniDeploymentView'
import type { TBranchCardData } from './BranchCard'

const mainGroupInstalls: TMiniDeployInstall[] = [
  ...[
    { name: 'staging-example', health: 'healthy', rolledOut: true },
    { name: 'staging-eu', health: 'degraded', rolledOut: true },
  ].map((install) => ({
    ...install,
    id: `install-${install.name}`,
    group: 'staging',
    href: `/org-1/installs/install-${install.name}`,
    runStatus: 'success',
  })),
  ...[
    { name: 'acme-prod', health: 'healthy', rolledOut: true },
    { name: 'globex-prod', health: 'healthy', rolledOut: true },
    { name: 'initech-prod', health: 'degraded', rolledOut: true },
    { name: 'umbrella-prod', health: 'healthy', rolledOut: false },
  ].map((install) => ({
    ...install,
    id: `install-${install.name}`,
    group: 'customers',
    href: `/org-1/installs/install-${install.name}`,
    runStatus: install.rolledOut ? 'success' : 'in-progress',
  })),
  ...[
    { name: 'soylent-prod', health: 'healthy', rolledOut: true },
    { name: 'hooli-prod', health: 'unhealthy', rolledOut: false },
    { name: 'vandelay-prod', health: 'healthy', rolledOut: true },
    { name: 'wonka-prod', health: 'unknown', rolledOut: false },
    { name: 'massive-prod', health: 'healthy', rolledOut: true },
    { name: 'vehement-prod', health: 'degraded', rolledOut: true },
    { name: 'stark-prod', health: 'healthy', rolledOut: true },
    { name: 'wayne-prod', health: 'healthy', rolledOut: true },
    { name: 'cyberdyne-prod', health: 'healthy', rolledOut: false },
    { name: 'tyrell-prod', health: 'healthy', rolledOut: true },
  ].map((install) => ({
    ...install,
    id: `install-${install.name}`,
    group: 'enterprise',
    href: `/org-1/installs/install-${install.name}`,
    runStatus: install.rolledOut
      ? 'success'
      : install.health === 'unhealthy'
        ? 'failed'
        : 'awaiting-approval',
  })),
]

export const mockBranchCards: TBranchCardData[] = [
  {
    branchId: 'br-001',
    name: 'main',
    href: '/org-1/apps/app-1/branches/br-001',
    repo: 'acme/platform-configs',
    repoBranch: 'main',
    latestRun: {
      href: '/org-1/apps/app-1/branches/br-001/runs/run-1',
      status: 'success',
      commitMessage: 'bump api image to v1.42.0',
      sha: '85d067ecafe1234',
      author: 'Example Developer',
      createdAt: '2026-07-30T10:30:00Z',
      trigger: 'tag',
      tag: 'v1.42.0',
    },
    planGroups: [
      { name: 'staging', installs: 2, hasSelector: false },
      { name: 'customers', installs: 4, hasSelector: false },
      { name: 'enterprise', installs: 10, hasSelector: false },
    ],
    latestRunInstalls: mainGroupInstalls,
  },
  {
    branchId: 'br-002',
    name: 'staging',
    href: '/org-1/apps/app-1/branches/br-002',
    repo: 'acme/platform-configs',
    repoBranch: 'staging',
    latestRun: {
      href: '/org-1/apps/app-1/branches/br-002/runs/run-2',
      status: 'running',
      commitMessage: 'add new worker component',
      sha: '6987a43568abc82',
      author: 'Example Developer',
      createdAt: '2026-08-01T14:00:00Z',
      awaitingApproval: true,
      trigger: 'pull_request',
      prNumber: 142,
    },
    planGroups: [
      { name: 'customers', installs: 2, hasSelector: true },
      { name: 'pr-previews', installs: 2, hasSelector: false, isPreview: true },
    ],
    latestRunInstalls: [
      {
        id: 'install-acme-staging',
        name: 'acme-staging',
        group: 'customers',
        href: '/org-1/installs/install-acme-staging',
        runStatus: 'in-progress',
        health: 'healthy',
        rolledOut: false,
      },
      {
        id: 'install-globex-staging',
        name: 'globex-staging',
        group: 'customers',
        href: '/org-1/installs/install-globex-staging',
        runStatus: 'awaiting-approval',
        health: 'degraded',
        rolledOut: false,
      },
      {
        id: 'install-preview-142',
        name: 'preview-142',
        group: 'pr-previews',
        href: '/org-1/installs/install-preview-142',
        runStatus: 'success',
        health: 'healthy',
        rolledOut: true,
      },
      {
        id: 'install-preview-143',
        name: 'preview-143',
        group: 'pr-previews',
        href: '/org-1/installs/install-preview-143',
        runStatus: 'in-progress',
        health: 'unknown',
        rolledOut: false,
      },
    ],
  },
  {
    branchId: 'br-003',
    name: 'feature/new-deploy',
    href: '/org-1/apps/app-1/branches/br-003',
    planGroups: [],
  },
]

export const manyBranchCards: TBranchCardData[] = [
  ...mockBranchCards,
  {
    ...mockBranchCards[0],
    branchId: 'br-004',
    name: 'release/2026-09',
    href: '/org-1/apps/app-1/branches/br-004',
  },
  {
    ...mockBranchCards[1],
    branchId: 'br-005',
    name: 'feature/usage-metrics',
    href: '/org-1/apps/app-1/branches/br-005',
    latestRun: {
      ...mockBranchCards[1].latestRun!,
      status: 'failed',
      awaitingApproval: false,
      trigger: 'manual',
      prNumber: undefined,
      tag: undefined,
    },
  },
  {
    ...mockBranchCards[2],
    branchId: 'br-006',
    name: 'customer/acme',
    href: '/org-1/apps/app-1/branches/br-006',
  },
]
