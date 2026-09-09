import { Outlet } from 'react-router'
import { Text } from '../components/atoms/Text'
import { BranchProfile } from '../components/molecules/BranchProfile'
import { SubNav, type ISubNavItem } from '../components/molecules/SubNav'
import { BranchSwitcher } from '../components/organisms/BranchSwitcher'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { usePageTitle } from '../hooks/use-page-title'
import { useStatusBar } from '../hooks/use-status-bar'
import type { IBreadcrumbItem } from '../providers/breadcrumb-provider'
import {
  AppBranchProvider,
  useAppBranch,
} from '../providers/app-branch-provider'
import { useApp } from '../providers/app-provider'
import { useOrg } from '../providers/org-provider'

export const appBranchNavigation = (
  orgId: string,
  appId: string,
  branchId: string
): ISubNavItem[] => {
  const base = `/${orgId}/apps/${appId}/branches/${branchId}`

  return [
    { href: base, label: 'Overview', end: true },
    { href: `${base}/activity`, label: 'Activity' },
    { href: `${base}/config`, label: 'Config' },
  ]
}

export const useAppBranchPageChrome = (section?: string) => {
  const { org, orgId } = useOrg()
  const { app, appId } = useApp()
  const { branchId } = useAppBranch()
  const overviewHref =
    orgId && appId && branchId
      ? `/${orgId}/apps/${appId}/branches/${branchId}`
      : undefined

  const trail: IBreadcrumbItem[] = [
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    {
      label: 'Apps',
      href: orgId ? `/${orgId}/apps` : undefined,
    },
    {
      label: app?.name,
      href: section ? overviewHref : undefined,
      loadingWidth: 14,
    },
  ]
  if (section) trail.push({ label: section })

  usePageTitle(section ?? app?.name, section ? app?.name : undefined)
  useBreadcrumbs(trail)
}

const AppBranchChrome = () => {
  const { orgId } = useOrg()
  const { app, appId, loading: appLoading } = useApp()
  const { branch, branchId, loading } = useAppBranch()

  useStatusBar(
    <span className="flex min-w-0 items-center gap-2">
      <Text
        variant="caption"
        family="mono"
        weight="medium"
        loading={appLoading}
        loadingWidth={12}
        className="truncate leading-tight"
      >
        {app?.name ?? 'App unavailable'}
      </Text>
      <Text variant="caption" color="tertiary" aria-hidden>
        /
      </Text>
      <BranchProfile branch={branch} loading={loading} variant="modeline" />
    </span>,
    [app?.name, appLoading, branch?.name, loading]
  )

  return (
    <div className="flex w-full flex-col gap-6">
      <div className="flex items-center justify-between gap-4">
        <SubNav
          items={appBranchNavigation(orgId ?? '', appId ?? '', branchId ?? '')}
          label="App sections"
        />
        <BranchSwitcher />
      </div>
      <Outlet />
    </div>
  )
}

export const AppBranchLayout = () => (
  <AppBranchProvider>
    <AppBranchChrome />
  </AppBranchProvider>
)
