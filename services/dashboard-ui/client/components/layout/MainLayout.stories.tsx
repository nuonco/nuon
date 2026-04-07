export default {
  title: 'Layout/MainLayout',
}

import type { ReactNode } from 'react'
import { SidebarContext } from '@/providers/sidebar-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { MainLayout } from './MainLayout'

const mockVersions = {
  api: { git_ref: 'abc1234', version: '1.2.3' },
  ui: { version: '4.5.6' },
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

const mockSidebarOpen = {
  isSidebarOpen: true,
  closeSidebar: () => {},
  openSidebar: () => {},
  toggleSidebar: () => {},
}

const mockSidebarClosed = {
  isSidebarOpen: false,
  closeSidebar: () => {},
  openSidebar: () => {},
  toggleSidebar: () => {},
}

const Providers = ({ sidebar, children }: { sidebar: typeof mockSidebarOpen; children: ReactNode }) => (
  <NotificationContext.Provider value={mockNotifications}>
    <SidebarContext.Provider value={sidebar}>
      {children}
    </SidebarContext.Provider>
  </NotificationContext.Provider>
)

export const SidebarOpen = () => (
  <Providers sidebar={mockSidebarOpen}>
    <MainLayout versions={mockVersions}>
      <div className="p-8">Page content</div>
    </MainLayout>
  </Providers>
)

export const SidebarClosed = () => (
  <Providers sidebar={mockSidebarClosed}>
    <MainLayout versions={mockVersions}>
      <div className="p-8">Page content</div>
    </MainLayout>
  </Providers>
)

export const HideOrgContent = () => (
  <Providers sidebar={mockSidebarOpen}>
    <MainLayout versions={mockVersions} hideOrgContent>
      <div className="p-8">Page content without org nav</div>
    </MainLayout>
  </Providers>
)
