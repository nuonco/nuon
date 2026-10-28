import type { ReactNode } from 'react'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { PageSidebarProvider } from '@/providers/page-sidebar-provider'
import { SidebarContext } from '@/providers/sidebar-provider'

const mockBreadcrumb = {
  breadcrumbLinks: [
    { path: '/org-1', text: 'Acme' },
    { path: '/org-1/shared-installs', text: 'Shared installs' },
  ],
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

export const SharedInstallsStoryChrome = ({
  children,
  breadcrumbLinks = mockBreadcrumb.breadcrumbLinks,
}: {
  children: ReactNode
  breadcrumbLinks?: typeof mockBreadcrumb.breadcrumbLinks
}) => (
  <NotificationContext.Provider value={mockNotifications}>
    <BreadcrumbContext.Provider
      value={{ ...mockBreadcrumb, breadcrumbLinks }}
    >
      <SidebarContext.Provider value={mockSidebar}>
        <PageSidebarProvider>{children}</PageSidebarProvider>
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)
