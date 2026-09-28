import type { TNavLink } from '@/types'

export type TInstallHrefInput = {
  orgId?: string
  appId?: string
  installId?: string
  suffix?: string
  nested?: boolean
}

export const isNewInstallIAEnabled = (
  features?: { [key: string]: boolean } | null
) => !!features?.['app-branches-ui'] && !!features?.['new-install-ia']

export const installHref = ({
  orgId,
  appId,
  installId,
  suffix = '',
  nested = false,
}: TInstallHrefInput) => {
  if (!orgId || !installId) return ''
  const base =
    nested && appId
      ? `/${orgId}/apps/${appId}/installs/${installId}`
      : `/${orgId}/installs/${installId}`
  if (!suffix) return base
  if (suffix.startsWith('/') || suffix.startsWith('?') || suffix.startsWith('#')) {
    return `${base}${suffix}`
  }
  return `${base}/${suffix}`
}

export const installPathnameSuffix = (pathname: string, installId: string) => {
  const marker = `/installs/${installId}`
  const index = pathname.indexOf(marker)
  if (index === -1) return ''
  return pathname.slice(index + marker.length)
}

export const withAppInstallBreadcrumbs = (
  crumbs: TNavLink[],
  {
    nested,
    orgId,
    appId,
    appName,
    installId,
  }: {
    nested: boolean
    orgId?: string
    appId?: string
    appName?: string
    installId?: string
  }
): TNavLink[] => {
  if (!nested || !orgId || !appId || !installId) return crumbs
  const listPath = `/${orgId}/installs`
  const root = `/${orgId}/installs/${installId}`
  const nestedRoot = installHref({ orgId, appId, installId, nested: true })
  const mentionsInstall = crumbs.some(
    (crumb) => crumb.path === root || crumb.path.startsWith(`${root}/`)
  )
  if (!mentionsInstall) return crumbs

  const rewritten = crumbs
    .filter((crumb) => crumb.path !== listPath)
    .map((crumb) =>
      crumb.path === root || crumb.path.startsWith(`${root}/`)
        ? { ...crumb, path: `${nestedRoot}${crumb.path.slice(root.length)}` }
        : crumb
    )
  const installIndex = rewritten.findIndex((crumb) => crumb.path === nestedRoot)
  const appCrumbs: TNavLink[] = [
    { path: `/${orgId}/apps`, text: 'Apps' },
    { path: `/${orgId}/apps/${appId}`, text: appName ?? '' },
  ]
  if (installIndex === -1) return [...appCrumbs, ...rewritten]
  return [
    ...rewritten.slice(0, installIndex),
    ...appCrumbs,
    ...rewritten.slice(installIndex),
  ]
}

export const installBreadcrumbs = ({
  nested,
  orgId,
  orgName,
  appId,
  appName,
  installId,
  installName,
  legacy,
  tail = [],
}: {
  nested: boolean
  orgId?: string
  orgName?: string
  appId?: string
  appName?: string
  installId?: string
  installName?: string
  legacy: TNavLink[]
  tail?: { suffix: string; text?: string }[]
}): TNavLink[] => {
  if (!nested || !appId) return legacy
  return [
    { path: `/${orgId}`, text: orgName ?? '' },
    { path: `/${orgId}/apps`, text: 'Apps' },
    { path: `/${orgId}/apps/${appId}`, text: appName ?? '' },
    {
      path: installHref({ orgId, appId, installId, nested: true }),
      text: installName ?? '',
    },
    ...tail.map((item) => ({
      path: installHref({
        orgId,
        appId,
        installId,
        nested: true,
        suffix: item.suffix,
      }),
      text: item.text ?? '',
    })),
  ]
}
