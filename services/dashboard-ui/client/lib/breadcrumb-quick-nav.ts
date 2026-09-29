import type { TNavLink } from '@/types'

type TQuickNavTarget = {
  orgId?: string
  appId?: string
  resourceId?: string
  name?: string
}

const branchSections = new Set([
  'actions',
  'components',
  'configs',
  'inputs',
  'install-configs',
  'installs',
  'labels',
  'plan',
  'policies',
  'readme',
  'roles',
  'rollout',
  'runbooks',
  'runs',
  'sandbox',
  'settings',
])

const branchStaticSuffixes = new Set([
  '',
  ...Array.from(branchSections, (section) => `/${section}`),
  '/policies/analytics',
])

const installStaticSuffixes = new Set([
  '',
  '/health',
  '/resources',
  '/resources/sandbox',
  '/resources/components',
  '/resources/images',
  '/operations',
  '/operations/actions',
  '/operations/runbooks',
  '/operations/policies',
  '/operations/runner',
  '/configuration',
  '/configuration/inputs',
  '/configuration/config-file',
  '/configuration/overrides',
  '/configuration/state',
])

const normalizedSuffix = (pathname: string, basePath: string) => {
  if (pathname !== basePath && !pathname.startsWith(`${basePath}/`)) return ''
  return pathname.slice(basePath.length).replace(/\/$/, '')
}

export const branchQuickNavHref = ({
  pathname,
  currentBasePath,
  targetBasePath,
}: {
  pathname: string
  currentBasePath: string
  targetBasePath: string
}) => {
  const suffix = normalizedSuffix(pathname, currentBasePath)
  if (branchStaticSuffixes.has(suffix)) return `${targetBasePath}${suffix}`

  const section = suffix.split('/').filter(Boolean)[0]
  return section && branchSections.has(section)
    ? `${targetBasePath}/${section}`
    : targetBasePath
}

const withQuickNav = (
  crumbs: TNavLink[],
  {
    enabled,
    type,
    target,
    basePath,
  }: {
    enabled: boolean
    type: 'branch' | 'install'
    target: TQuickNavTarget
    basePath: string
  }
) => {
  const { orgId, appId, resourceId, name } = target
  if (!enabled || !orgId || !appId || !resourceId || !name) return crumbs

  let attached = false
  return crumbs.map((crumb) => {
    if (attached || crumb.path !== basePath || crumb.text !== name) return crumb
    attached = true
    return {
      ...crumb,
      quickNav: {
        type,
        orgId,
        appId,
        resourceId,
        basePath,
      },
    }
  })
}

export const withBreadcrumbQuickNav = (
  crumbs: TNavLink[],
  {
    branch,
    install,
  }: {
    branch?: TQuickNavTarget & { enabled: boolean }
    install?: TQuickNavTarget & { enabled: boolean }
  }
) => {
  const withBranch = withQuickNav(crumbs, {
    enabled: !!branch?.enabled,
    type: 'branch',
    target: branch ?? {},
    basePath:
      branch?.orgId && branch.appId && branch.resourceId
        ? `/${branch.orgId}/apps/${branch.appId}/branches/${branch.resourceId}`
        : '',
  })
  return withQuickNav(withBranch, {
    enabled: !!install?.enabled,
    type: 'install',
    target: install ?? {},
    basePath:
      install?.orgId && install.appId && install.resourceId
        ? `/${install.orgId}/apps/${install.appId}/installs/${install.resourceId}`
        : '',
  })
}

export const installQuickNavHref = ({
  pathname,
  currentBasePath,
  targetBasePath,
}: {
  pathname: string
  currentBasePath: string
  targetBasePath: string
}) => {
  const suffix = normalizedSuffix(pathname, currentBasePath)
  return installStaticSuffixes.has(suffix)
    ? `${targetBasePath}${suffix}`
    : targetBasePath
}
