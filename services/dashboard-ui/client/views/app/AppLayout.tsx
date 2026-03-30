import { Outlet, useParams, useLocation } from 'react-router'
import { TemporalLink } from '@/components/admin/TemporalLink'
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
  const location = useLocation()
  const { org } = useOrg()
  const { app } = useApp()
  const isThirdLevel = location.pathname.split('/').length > 5
  const hasAppBranches = !!org?.features?.['app-branches']

  if (!app) return null

  const navLinks = [
    { path: `/`, iconVariant: 'HouseSimple' as const, text: 'Overview' },
    hasAppBranches && {
      path: `/sandbox`,
      iconVariant: 'ShippingContainerIcon' as const,
      text: 'Sandbox builds',
    },
    { path: `/components`, iconVariant: 'Cards' as const, text: 'Components' },
    { path: `/actions`, iconVariant: 'TerminalWindow' as const, text: 'Actions' },
    hasAppBranches && {
      path: `/branches`,
      iconVariant: 'GitBranch' as const,
      text: 'Branches',
    },
    { path: `/roles`, iconVariant: 'FileLock' as const, text: 'Roles' },
    { path: `/policies`, iconVariant: 'ShieldCheck' as const, text: 'Policies' },
    { path: `/installs`, iconVariant: 'Cube' as const, text: 'Installs' },
    { path: `/readme`, iconVariant: 'BookOpen' as const, text: 'README' },
  ].filter(Boolean) as TNavLink[]

  return (
    <PageLayout>
      {isThirdLevel ? (
        <PageContent className="border-t" isScrollable variant="secondary">
          <SubNav
            basePath={`/${org?.id}/apps/${app?.id}`}
            links={navLinks}
          />
          <div className="flex flex-col w-full">
            <Outlet />
          </div>
        </PageContent>
      ) : (
        <>
          <PageHeader>
            <PageHeadingGroup title={app.name} subtitle={<ID>{app.id}</ID>} />
            <div className="flex items-center gap-4">
              <TemporalLink namespace="apps" eventLoopId={app?.id} />
              {app?.runner_config ? (
                <CreateInstallButton variant="primary" />
              ) : null}
            </div>
          </PageHeader>
          <PageContent className="border-t" isScrollable variant="secondary">
            <SubNav
              basePath={`/${org?.id}/apps/${app?.id}`}
              links={navLinks}
            />
            <Outlet />
          </PageContent>
        </>
      )}
    </PageLayout>
  )
}
