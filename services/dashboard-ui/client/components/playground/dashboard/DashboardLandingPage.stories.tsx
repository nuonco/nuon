export default {
  title: 'Playground/Dashboard/LandingPage',
  fullBleed: true,
}

import type { ReactNode } from 'react'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { DashboardLandingPage } from './DashboardLandingPage'
import {
  dashboardLandingAnnouncements,
  dashboardLandingRuns,
} from './DashboardLandingPage.fixtures'

const mockBreadcrumb = {
  breadcrumbLinks: [{ path: '/org-1', text: 'Acme' }],
  isLoading: false,
  updateBreadcrumb: () => {},
}

const mockSidebar = {
  isSidebarOpen: true,
  closeSidebar: () => {},
  openSidebar: () => {},
  toggleSidebar: () => {},
}

const mockNotifications = {
  emitNotification: async () => false,
  permission: 'default' as NotificationPermission,
  requestPermission: async () => 'default' as NotificationPermission,
  isSupported: false,
  settings: { permissionRequested: false },
  hasRequestedPermission: false,
  muted: false,
  toggleMute: () => {},
}

const Providers = ({ children }: { children: ReactNode }) => (
  <NotificationContext.Provider value={mockNotifications}>
    <BreadcrumbContext.Provider value={mockBreadcrumb}>
      <SidebarContext.Provider value={mockSidebar}>
        {children}
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)

export const Default = () => (
  <Providers>
    <div className="flex h-full min-h-0">
      <DashboardLandingPage
        orgName="Acme"
        runs={dashboardLandingRuns}
        announcements={dashboardLandingAnnouncements}
      />
    </div>
  </Providers>
)

export const Quiet = () => (
  <Providers>
    <div className="flex h-full min-h-0">
      <DashboardLandingPage
        orgName="Acme"
        runs={[]}
        announcements={dashboardLandingAnnouncements}
      />
    </div>
  </Providers>
)
