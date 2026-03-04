import { Outlet } from 'react-router'
import { MainLayout } from '@/components/layout/MainLayout'
import { APIHealthProvider } from '@/providers/api-health-provider'
import { BreadcrumbProvider } from '@/providers/breadcrumb-provider'
import { NotificationProvider } from '@/providers/notification-provider'
import { OrgProvider } from '@/providers/org-provider'
import { SidebarProvider } from '@/providers/sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { getSidebarOpen } from '@/lib/cookies'

export const OrgLayout = () => {
  return (
    <NotificationProvider autoRequestOnLoad={true} autoRequestDelay={3000}>
      {/* <APIHealthProvider shouldPoll> */}
      <OrgProvider>
        <BreadcrumbProvider>
          <SidebarProvider initIsSidebarOpen={getSidebarOpen()}>
            <ToastProvider>
              <SurfacesProvider>
                <MainLayout
                  versions={{
                    api: {
                      git_ref: '',
                      version: 'local',
                    },
                    ui: {
                      version: 'local',
                    },
                  }}
                >
                  <Outlet />
                </MainLayout>
              </SurfacesProvider>
            </ToastProvider>
          </SidebarProvider>
        </BreadcrumbProvider>
      </OrgProvider>
      {/* </APIHealthProvider> */}
    </NotificationProvider>
  )
}
