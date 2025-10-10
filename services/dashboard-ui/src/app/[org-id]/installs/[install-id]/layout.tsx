import { notFound } from 'next/navigation'
import { getIsPageSidebarOpenFromCookie } from '@/actions/layout/page-sidebar-cookie'
import { getInstallById, getOrgById } from '@/lib'
import { PageSidebarProvider } from '@/providers/page-sidebar-provider'
import { InstallProvider } from '@/providers/install-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import type { TLayoutProps } from '@/types'

interface IInstallLayout extends TLayoutProps<'org-id' | 'install-id'> {}

export default async function InstallLayout({
  children,
  params,
}: IInstallLayout) {
  const isPageSidebarOpen = await getIsPageSidebarOpenFromCookie()
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const [{ data: install, error }, { data: org }] = await Promise.all([
    getInstallById({ installId, orgId }),
    getOrgById({ orgId }),
  ])

  if (error) {
    console.error('error fetching install by id', error)
    notFound()
  }

  return (
    <InstallProvider initInstall={install} shouldPoll>
      {org?.features?.['stratus-layout'] ? (
        <PageSidebarProvider initIsPageSidebarOpen={isPageSidebarOpen}>
          <ToastProvider>
            <SurfacesProvider>{children}</SurfacesProvider>
          </ToastProvider>
        </PageSidebarProvider>
      ) : (
        children
      )}
    </InstallProvider>
  )
}
