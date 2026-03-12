import { useQuery } from '@tanstack/react-query'
import { Outlet } from 'react-router'
import { MainLayout } from '@/components/layout/MainLayout'
import { useConfig } from '@/hooks/use-config'
import { getSidebarOpen } from '@/lib/cookies'
import { getAPIVersion } from '@/lib'
import { APIHealthProvider } from '@/providers/api-health-provider'
import { BreadcrumbProvider } from '@/providers/breadcrumb-provider'
import { PageTitleProvider } from '@/providers/page-title-provider'
import { NotificationProvider } from '@/providers/notification-provider'
import { OrgProvider } from '@/providers/org-provider'
import { SidebarProvider } from '@/providers/sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'

export const OrgLayout = () => {
  const { version } = useConfig()
  const { data: apiVersion } = useQuery({
    queryKey: ['api-version'],
    queryFn: getAPIVersion,
  })

  return (
    <NotificationProvider autoRequestOnLoad={true} autoRequestDelay={3000}>
      <APIHealthProvider shouldPoll>
        <OrgProvider>
          <BreadcrumbProvider>
            <PageTitleProvider>
              <SidebarProvider initIsSidebarOpen={getSidebarOpen()}>
                <ToastProvider>
                  <SurfacesProvider>
                    <MainLayout
                      versions={{
                        api: {
                          git_ref: apiVersion?.git_ref ?? '',
                          version: apiVersion?.version ?? '',
                        },
                        ui: {
                          version: version,
                        },
                      }}
                    >
                      <Outlet />
                    </MainLayout>
                  </SurfacesProvider>
                </ToastProvider>
              </SidebarProvider>
            </PageTitleProvider>
          </BreadcrumbProvider>
        </OrgProvider>
      </APIHealthProvider>
    </NotificationProvider>
  )
}
