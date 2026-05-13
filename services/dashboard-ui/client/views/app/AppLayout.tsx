import { Outlet, useParams, useMatch } from 'react-router'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { CreateInstallButton } from '@/components/apps/CreateInstall'
import { ID } from '@/components/common/ID'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageHeadingGroup } from '@/components/layout/PageHeadingGroup'
import { SubNav } from '@/components/navigation/SubNav'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { AppProvider } from '@/providers/app-provider'
import { PageSidebarProvider } from '@/providers/page-sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import type { TNavLink } from '@/types/dashboard.types'

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

const AppTemplate = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const isChildRoute = !!useMatch('/:orgId/apps/:appId/:section/:rest/*')
  const hasAppBranchesUI = !!org?.features?.['app-branches-ui']

  if (!app) return null

  const navLinks = [
    { path: `/`, iconVariant: 'HouseSimpleIcon' as const, text: 'Overview' },
    hasAppBranchesUI && {
      path: `/sandbox`,
      iconVariant: 'ShippingContainerIcon' as const,
      text: 'Sandbox builds',
    },
    { path: `/components`, iconVariant: 'CardsIcon' as const, text: 'Components' },
    { path: `/actions`, iconVariant: 'TerminalWindowIcon' as const, text: 'Actions' },
    hasAppBranchesUI && {
      path: `/branches`,
      iconVariant: 'GitBranchIcon' as const,
      text: 'Branches',
    },
    { path: `/roles`, iconVariant: 'FileLockIcon' as const, text: 'Roles' },
    { path: `/policies`, iconVariant: 'ShieldCheckIcon' as const, text: 'Policies' },
    { path: `/installs`, iconVariant: 'CubeIcon' as const, text: 'Installs' },
    { path: `/readme`, iconVariant: 'BookOpenIcon' as const, text: 'README' },
  ].filter(Boolean) as TNavLink[]

  return (
    <PageLayout>
      {isChildRoute ? (
        <PageContent className="border-t" variant="row">
          <SubNav
            basePath={`/${org?.id}/apps/${app?.id}`}
            links={navLinks}
          />
          <div className="flex flex-col flex-1 min-w-0">
            <Outlet />
          </div>
        </PageContent>
      ) : (
        <>
          <PageHeader>
            <div className="flex flex-col gap-4 w-full md:flex-row md:justify-between md:items-start">
              <PageHeadingGroup title={app.name} subtitle={<ID>{app.id}</ID>} />
              <div className="flex items-center gap-4">
                <AdminDashboardLink path={`/queues?owner_id=${app.id}&owner_type=apps`} label="View queues" />
                {app?.runner_config ? (
                  <CreateInstallButton variant="primary" />
                ) : null}
              </div>
            </div>
          </PageHeader>
          <PageContent className="border-t" variant="row">
            <SubNav
              basePath={`/${org?.id}/apps/${app?.id}`}
              links={navLinks}
            />
             <div className="flex flex-col flex-1 min-w-0">
               <Outlet />
             </div>
          </PageContent>
        </>
      )}
    </PageLayout>
  )
}
