import type { INavItem } from './types'

export const primaryNav: INavItem[] = [
  { label: 'Dashboard', path: '/', icon: 'SquaresFourIcon' },
  { label: 'Apps', path: '/apps', icon: 'AppWindowIcon' },
  { label: 'Installs', path: '/installs', icon: 'PackageIcon' },
]

export const secondaryNav: INavItem[] = [
  { label: 'Team', path: '/team', icon: 'UsersIcon' },
  { label: 'Docs', path: '/docs', icon: 'BookOpenIcon' },
  { label: 'Settings', path: '/settings', icon: 'GearIcon' },
]

export const branchBase = (appId: string, branchId: string) =>
  `/apps/${appId}/branches/${branchId}`

export const appTabs = (appId: string, branchId: string): INavItem[] => {
  const base = branchBase(appId, branchId)

  return [
    { label: 'Overview', path: base },
    { label: 'Activity', path: `${base}/activity` },
    { label: 'Config', path: `${base}/config` },
  ]
}


export const installTabs = (installId: string): INavItem[] => [
  { label: 'Overview', path: `/installs/${installId}` },
  { label: 'Activity', path: `/installs/${installId}/activity` },
]

export const runTabs = (basePath: string): INavItem[] => [
  { label: 'Summary', path: basePath },
  { label: 'Logs', path: `${basePath}/logs` },
  { label: 'Trace', path: `${basePath}/trace` },
  { label: 'Outputs', path: `${basePath}/outputs` },
]

export const settingsNav: INavItem[] = [
  { label: 'Connections', path: '/settings', icon: 'GitHub' },
  { label: 'Webhooks', path: '/settings/webhooks', icon: 'WebhooksLogoIcon' },
  { label: 'Triggers', path: '/settings/triggers', icon: 'LightningIcon' },
  { label: 'API tokens', path: '/settings/api-tokens', icon: 'KeyIcon' },
  {
    label: 'Service accounts',
    path: '/settings/service-accounts',
    icon: 'RobotIcon',
  },
  { label: 'OIDC federation', path: '/settings/oidc', icon: 'ShieldCheckIcon' },
]
