import { notFound } from 'next/navigation'
import { getIsSidebarOpenFromCookie } from '@/actions/layout/main-sidebar-cookie'
import { Layout as OldLayout } from '@/components/Layout'
import { MainLayout } from '@/components/layout/MainLayout'
import { REFRESH_PAGE_INTERVAL, REFRESH_PAGE_WARNING } from '@/configs/app'
import { getAPIVersion, getOrgById, getOrgs } from '@/lib'
import { AutoRefreshProvider } from '@/providers/auto-refresh-provider'
import { OrgProvider } from '@/providers/org-provider'
import { SidebarProvider } from '@/providers/sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import type { TLayoutProps } from '@/types'
import { VERSION } from '@/utils'

export default async function OrgLayout({
  children,
  params,
}: TLayoutProps<'org-id'>) {
  const isSidebarOpen = await getIsSidebarOpenFromCookie()
  const { ['org-id']: orgId } = await params
  const [{ data: org, error }, { data: orgs }, { data: apiVersion }] =
    await Promise.all([
      getOrgById({ orgId }).catch((error) => {
        console.error(error)
        notFound()
      }),
      getOrgs().catch((error) => {
        console.error(error)
        notFound()
      }),
      getAPIVersion(),
    ])

  if (error) {
    notFound()
  }

  return (
    <AutoRefreshProvider
      refreshIntervalMs={REFRESH_PAGE_INTERVAL as number}
      showWarning={REFRESH_PAGE_WARNING as boolean}
      warningTimeMs={30 * 1000} // 30 second warning
    >
      <OrgProvider initOrg={org} shouldPoll>
        {org?.features?.['stratus-layout'] ? (
          <SidebarProvider initIsSidebarOpen={isSidebarOpen}>
            <ToastProvider>
              <SurfacesProvider>
                <MainLayout>{children}</MainLayout>
              </SurfacesProvider>
            </ToastProvider>
          </SidebarProvider>
        ) : (
          <OldLayout
            isSidebarOpen={isSidebarOpen}
            orgs={orgs}
            versions={
              {
                api: apiVersion,
                ui: {
                  version: VERSION,
                },
              } as any
            }
          >
            {children}
          </OldLayout>
        )}
      </OrgProvider>
    </AutoRefreshProvider>
  )
}
