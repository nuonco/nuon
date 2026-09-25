import type { TBranchPlanGroup } from '@/components/branches/BranchCards/BranchPlanDots'
import type { TMiniDeployInstall } from './MiniDeploymentView'

export const miniDeployGroups: TBranchPlanGroup[] = [
  { name: 'canary', installs: 2, hasSelector: false },
  { name: 'enterprise', installs: 10, hasSelector: false },
  { name: 'rest', installs: 3, hasSelector: true },
]

const enterpriseNames = [
  'production-acme',
  'production-globex',
  'production-initech',
  'production-umbrella',
  'production-soylent',
  'production-hooli',
  'production-vehement',
  'production-massive',
  'production-vandelay',
  'production-wonka',
]

const enterpriseRunStatuses = [
  'success',
  'success',
  'success',
  'in-progress',
  'in-progress',
  'failed',
  'success',
  'awaiting-approval',
  'success',
  'success',
]

const enterpriseHealth = [
  'healthy',
  'healthy',
  'degraded',
  'healthy',
  'unknown',
  'unhealthy',
  'healthy',
  'healthy',
  'degraded',
  'healthy',
]

export const miniDeployInstalls: TMiniDeployInstall[] = [
  {
    id: 'install-staging-example',
    name: 'staging-example',
    group: 'canary',
    href: '/org-1/installs/install-staging-example',
    runStatus: 'success',
    health: 'healthy',
    rolledOut: true,
  },
  {
    id: 'install-preview-1042',
    name: 'preview-1042',
    group: 'canary',
    href: '/org-1/installs/install-preview-1042',
    runStatus: 'success',
    health: 'degraded',
    rolledOut: true,
  },
  ...enterpriseNames.map((name, index) => ({
    id: `install-${name}`,
    name,
    group: 'enterprise',
    href: `/org-1/installs/install-${name}`,
    runStatus: enterpriseRunStatuses[index],
    health: enterpriseHealth[index],
    rolledOut: enterpriseRunStatuses[index] === 'success',
  })),
  {
    id: 'install-development',
    name: 'development',
    group: 'rest',
    href: '/org-1/installs/install-development',
    runStatus: 'queued',
    health: 'healthy',
    rolledOut: false,
  },
  {
    id: 'install-demo-eu',
    name: 'demo-eu',
    group: 'rest',
    href: '/org-1/installs/install-demo-eu',
    runStatus: 'queued',
    health: 'unknown',
    rolledOut: false,
  },
  {
    id: 'install-demo-us',
    name: 'demo-us',
    group: 'rest',
    href: '/org-1/installs/install-demo-us',
    runStatus: 'queued',
    health: 'healthy',
    rolledOut: false,
  },
]
