import { notFound } from 'next/navigation'
import { getIsPageSidebarOpenFromCookie } from '@/actions/layout/page-sidebar-cookie'
import { getAppById, getOrgById } from '@/lib'
import { AppProvider } from '@/providers/app-provider'
import { PageSidebarProvider } from '@/providers/page-sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import type { TLayoutProps } from '@/types'

interface IAppLayout extends TLayoutProps<'org-id' | 'app-id'> {}

export default async function AppLayout({ children, params }: IAppLayout) {
  const isPageSidebarOpen = await getIsPageSidebarOpenFromCookie()
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [{ data: app, error }, { data: org }] = await Promise.all([
    getAppById({
      orgId,
      appId,
    }),
    getOrgById({ orgId }),
  ])

  if (error) {
    console.error('error fetching app by id', error)
    notFound()
  }

  return (
    <AppProvider initApp={app} shouldPoll>
      {org?.features?.['stratus-layout'] ? (
        <PageSidebarProvider initIsPageSidebarOpen={isPageSidebarOpen}>
          <ToastProvider>
            <SurfacesProvider>{children}</SurfacesProvider>
          </ToastProvider>
        </PageSidebarProvider>
      ) : (
        children
      )}
    </AppProvider>
  )
}
