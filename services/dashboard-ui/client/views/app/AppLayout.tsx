import { Outlet, useParams, useMatch } from 'react-router'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { CreateInstallButton } from '@/components/apps/CreateInstall'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { SubNav } from '@/components/navigation/SubNav'
import { useApp } from '@/hooks/use-app'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { AppProvider } from '@/providers/app-provider'
import { PageSidebarProvider } from '@/providers/page-sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { AppSourceChip } from '@/components/apps/AppSourceChip'
import type { TNavItem } from '@/types/dashboard.types'

export const AppLayout = () => {
  const params = useParams()

  return (
    <AppProvider appId={params?.appId} shouldPoll>
      <PageSidebarProvider>
        <ToastProvider>
          <SurfacesProvider>
            <AppTemplate />
          </SurfacesProvider>
        </ToastProvider>
      </PageSidebarProvider>
    </AppProvider>
  )
}

const AppHeader = ({ isBranchPicker }: { isBranchPicker?: boolean }) => {
  const { app } = useApp()

  return (
    <DetailHeader
      variant="page"
      backLink={false}
      title={app?.name}
      id={app?.id}
      identity={isBranchPicker ? <AppSourceChip /> : null}
      actions={
        <>
          <AdminDashboardLink
            path={`/queues?owner_id=${app?.id}&owner_type=apps`}
            label="View queues"
          />
          {!isBranchPicker && app?.runner_config ? (
            <CreateInstallButton variant="primary" />
          ) : null}
        </>
      }
    />
  )
}

const AppTemplate = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const isChildRoute = !!useMatch('/:orgId/apps/:appId/:section/:rest/*')
  const hasAppBranchesUI = !!org?.features?.['app-branches-ui']
  const hasInstallSyncing = !!org?.features?.['app-install-syncing']
  const hasNewAppIA = useNewAppIA()
  const hasCustomerManagedInstalls =
    !!org?.features?.['customer-managed-installs']

  if (!app) return null

  if (hasNewAppIA) {
    return (
      <PageLayout>
        {!isChildRoute ? <AppHeader isBranchPicker /> : null}
        <Outlet />
      </PageLayout>
    )
  }

  const navLinks = [
    { path: `/`, iconVariant: 'HouseSimpleIcon' as const, text: 'Overview' },
    hasAppBranchesUI && {
      path: `/sandbox`,
      iconVariant: 'ShippingContainerIcon' as const,
      text: 'Sandbox builds',
    },
    {
      path: `/components`,
      iconVariant: 'CardsIcon' as const,
      text: 'Components',
    },
    {
      path: `/actions`,
      iconVariant: 'TerminalWindowIcon' as const,
      text: 'Actions',
    },
    {
      path: `/runbooks`,
      iconVariant: 'BookIcon' as const,
      text: 'Runbooks',
    },
    hasAppBranchesUI && {
      path: `/branches`,
      iconVariant: 'GitBranchIcon' as const,
      text: 'Branches',
    },
    { path: `/roles`, iconVariant: 'FileLockIcon' as const, text: 'Roles' },
    {
      path: `/policies`,
      iconVariant: 'ShieldCheckIcon' as const,
      text: 'Policies',
    },
    hasCustomerManagedInstalls && {
      path: `/releases`,
      iconVariant: 'PackageIcon' as const,
      text: 'Releases',
    },
    { path: `/installs`, iconVariant: 'CubeIcon' as const, text: 'Installs' },
    hasInstallSyncing && {
      path: `/install-syncs`,
      iconVariant: 'ArrowsClockwiseIcon' as const,
      text: 'Install syncs',
    },
    { path: `/labels`, iconVariant: 'TagIcon' as const, text: 'Labels' },
    { path: `/readme`, iconVariant: 'BookOpenIcon' as const, text: 'README' },
  ].filter(Boolean) as TNavItem[]

  return (
    <PageLayout>
      {!isChildRoute ? <AppHeader /> : null}
      <PageContent className="border-t" variant="row">
        <SubNav
          basePath={`/${org?.id}/apps/${app?.id}`}
          links={navLinks}
          storageKey="subnav:app"
        />
        <div className="flex flex-col flex-1 min-w-0">
          <Outlet />
        </div>
      </PageContent>
    </PageLayout>
  )
}
