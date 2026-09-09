import { Outlet } from 'react-router'
import { SubNav, type ISubNavItem } from '../components/molecules/SubNav'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import type { IBreadcrumbItem } from '../providers/breadcrumb-provider'
import { InstallProvider, useInstall } from '../providers/install-provider'
import { useOrg } from '../providers/org-provider'

export const installNavigation = (
  orgId: string,
  installId: string
): ISubNavItem[] => {
  const base = `/${orgId}/installs/${installId}`

  return [
    { href: base, label: 'Overview', end: true },
    { href: `${base}/activity`, label: 'Activity' },
  ]
}

export const useInstallPageChrome = (section?: string) => {
  const { org, orgId } = useOrg()
  const { install, installId } = useInstall()
  const overviewHref =
    orgId && installId ? `/${orgId}/installs/${installId}` : undefined

  const trail: IBreadcrumbItem[] = [
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    {
      label: 'Installs',
      href: orgId ? `/${orgId}/installs` : undefined,
    },
    {
      label: install?.name,
      href: section ? overviewHref : undefined,
      loadingWidth: 14,
    },
  ]
  if (section) trail.push({ label: section })

  usePageTitle(section ?? install?.name, section ? install?.name : undefined)
  useBreadcrumbs(trail)
}

const InstallChrome = () => {
  const { orgId } = useOrg()
  const { installId } = useInstall()

  return (
    <div className="flex w-full flex-col gap-6">
      <SubNav
        items={installNavigation(orgId ?? '', installId ?? '')}
        label="Install sections"
      />
      <Outlet />
    </div>
  )
}

export const InstallLayout = () => (
  <InstallProvider>
    <InstallChrome />
  </InstallProvider>
)
