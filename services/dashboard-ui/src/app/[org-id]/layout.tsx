import { notFound } from 'next/navigation'
import { getIsSidebarOpenFromCookie } from '@/actions/layout/main-sidebar-cookie'
import { MainLayout } from '@/components/layout/MainLayout'
import {
  REFRESH_PAGE_INTERVAL,
  REFRESH_PAGE_WARNING,
  VERSION,
} from '@/configs/app'
import { getAPIVersion, getOrgById, getOrgs } from '@/lib'
import { APIHealthProvider } from '@/providers/api-health-provider'
import { AutoRefreshProvider } from '@/providers/auto-refresh-provider'
import { BreadcrumbProvider } from '@/providers/breadcrumb-provider'
import { NotificationProvider } from '@/providers/notification-provider'
import { OrgProvider } from '@/providers/org-provider'
import { SidebarProvider } from '@/providers/sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import type { TLayoutProps } from '@/types'

// NOTE: old layout stuff
import { Layout as OldLayout } from '@/components/old/Layout'

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
    <NotificationProvider autoRequestOnLoad={true} autoRequestDelay={3000}>
      <APIHealthProvider shouldPoll>
        <AutoRefreshProvider
          refreshIntervalMs={REFRESH_PAGE_INTERVAL as number}
          showWarning={REFRESH_PAGE_WARNING as boolean}
          warningTimeMs={30 * 1000} // 30 second warning
        >
          <OrgProvider initOrg={org} shouldPoll>
            {org?.features?.['stratus-layout'] ? (
              <BreadcrumbProvider>
                <SidebarProvider initIsSidebarOpen={isSidebarOpen}>
                  <ToastProvider>
                    <SurfacesProvider>
                      <MainLayout
                        versions={{
                          api: apiVersion,
                          ui: {
                            version: VERSION,
                          },
                        }}
                      >
                        {children}
                      </MainLayout>
                    </SurfacesProvider>
                  </ToastProvider>
                </SidebarProvider>
              </BreadcrumbProvider>
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
      </APIHealthProvider>
    </NotificationProvider>
  )
}
