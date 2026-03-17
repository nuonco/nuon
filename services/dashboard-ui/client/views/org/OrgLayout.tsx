import { useQuery } from '@tanstack/react-query'
import { Outlet } from 'react-router'
import { MainLayout } from '@/components/layout/MainLayout'
import { OrgStatusBar } from '@/components/orgs/OrgStatusBar'
import { getSidebarOpen } from '@/lib/cookies'
import { getAPIVersion } from '@/lib'
import { APIHealthProvider } from '@/providers/api-health-provider'
import { BreadcrumbProvider } from '@/providers/breadcrumb-provider'
import { PageTitleProvider } from '@/providers/page-title-provider'
import { NotificationProvider } from '@/providers/notification-provider'
import { OrgProvider } from '@/providers/org-provider'
import { WorkflowApprovalsProvider } from '@/providers/workflow-approvals-provider'
import { SidebarProvider } from '@/providers/sidebar-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'

export const OrgLayout = () => {
  const { data: versions } = useQuery({
    queryKey: ['version'],
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
                  <WorkflowApprovalsProvider>
                    <SurfacesProvider>
                      <MainLayout
                        versions={{
                          api: {
                            git_ref: versions?.api?.git_ref ?? '',
                            version: versions?.api?.version ?? '',
                          },
                          ui: {
                            version: versions?.ui?.version ?? '',
                          },
                        }}
                      >
                        <Outlet />
                        <OrgStatusBar />
                      </MainLayout>
                    </SurfacesProvider>
                  </WorkflowApprovalsProvider>
                </ToastProvider>
              </SidebarProvider>
            </PageTitleProvider>
          </BreadcrumbProvider>
        </OrgProvider>
      </APIHealthProvider>
    </NotificationProvider>
  )
}
