import { Outlet, useParams } from 'react-router-dom'
import { setCookie } from '@/utils/cookies'
import { usePolling } from '@/hooks/use-polling'
import { NotificationProvider } from '@/providers/notification-provider'
import { APIHealthProvider } from '@/providers/api-health-provider'
import { AutoRefreshProvider } from '@/providers/auto-refresh-provider'
import { OrgContext } from '@/providers/org-provider'
import { BreadcrumbProvider } from '@/providers/breadcrumb-provider'
import { SidebarProvider } from '@/providers/sidebar-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { MainLayout } from '@/components/layout/MainLayout'
import type { TOrg } from '@/types'

const VERSION = process.env.VERSION || 'development'

export default function OrgLayout() {
  const { orgId } = useParams()

  const {
    data: org,
    error,
    isLoading,
  } = usePolling<TOrg>({
    path: `/api/ctl-api/v1/orgs/current`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  if (!org && error && !isLoading) {
    const errorMsg = error?.error || error?.description || error?.message || String(error)
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <p className="text-red-600 mb-4">Failed to load organization</p>
          <p className="text-gray-500 text-sm mb-4 max-w-md">{errorMsg}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  if (!org) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900" />
      </div>
    )
  }

  // Set org cookie for session persistence
  if (orgId) {
    setCookie('org_session', orgId, 365)
    setCookie('nuon-org-id', orgId, 365)
  }

  return (
    <NotificationProvider autoRequestOnLoad autoRequestDelay={3000}>
      <APIHealthProvider shouldPoll>
        <AutoRefreshProvider>
          <OrgContext.Provider
            value={{
              org: org || null,
              isLoading,
              error,
              refresh: () => {},
            }}
          >
            <BreadcrumbProvider>
              <SidebarProvider>
                <ToastProvider>
                  <SurfacesProvider>
                    <MainLayout
                      versions={{
                        api: { git_ref: '', version: '' },
                        ui: { version: VERSION },
                      }}
                    >
                      <Outlet />
                    </MainLayout>
                  </SurfacesProvider>
                </ToastProvider>
              </SidebarProvider>
            </BreadcrumbProvider>
          </OrgContext.Provider>
        </AutoRefreshProvider>
      </APIHealthProvider>
    </NotificationProvider>
  )
}
