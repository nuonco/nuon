export default {
  title: 'Playground / App Branches V3',
  fullBleed: true,
}

import type { ReactNode } from 'react'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { PageSidebarContext } from '@/providers/page-sidebar-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { AppBranchesV3 } from './AppBranchesV3'

const mockBreadcrumb = {
  breadcrumbLinks: [
    { path: '/org-1', text: 'Acme' },
    { path: '/org-1/apps', text: 'Apps' },
    { path: '/org-1/apps/app-1', text: 'acme-platform' },
    { path: '/org-1/apps/app-1/branches/brn-1', text: 'main' },
  ],
  isLoading: false,
  updateBreadcrumb: () => {},
}

const mockSidebar = {
  isSidebarOpen: false,
  closeSidebar: () => {},
  openSidebar: () => {},
  toggleSidebar: () => {},
}

const mockPageSidebar = {
  isPageSidebarOpen: true,
  closePageSidebar: () => {},
  openPageSidebar: () => {},
  togglePageSidebar: () => {},
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
        <PageSidebarContext.Provider value={mockPageSidebar}>
          {children}
        </PageSidebarContext.Provider>
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)

export const Default = () => (
  <Providers>
    <div className="flex h-full min-h-0">
      <AppBranchesV3 />
    </div>
  </Providers>
)
