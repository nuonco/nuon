export default {
  title: 'Layout/MainSidebar',
}

import { SidebarContext } from '@/providers/sidebar-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { MainSidebar } from './MainSidebar'

const mockVersions = {
  api: { git_ref: 'abc1234', version: '1.2.3' },
  ui: { version: '4.5.6' },
}

const noop = () => {}

const mockNotifications = {
  emitNotification: async () => false,
  permission: 'default' as NotificationPermission,
  requestPermission: async () => 'default' as NotificationPermission,
  isSupported: false,
  settings: { permissionRequested: false },
  hasRequestedPermission: false,
  muted: false,
  toggleMute: noop,
}

export const Open = () => (
  <NotificationContext.Provider value={mockNotifications}>
    <SidebarContext.Provider
      value={{ isSidebarOpen: true, closeSidebar: noop, openSidebar: noop, toggleSidebar: noop }}
    >
      <MainSidebar versions={mockVersions} />
    </SidebarContext.Provider>
  </NotificationContext.Provider>
)

export const Collapsed = () => (
  <NotificationContext.Provider value={mockNotifications}>
    <SidebarContext.Provider
      value={{ isSidebarOpen: false, closeSidebar: noop, openSidebar: noop, toggleSidebar: noop }}
    >
      <MainSidebar versions={mockVersions} />
    </SidebarContext.Provider>
  </NotificationContext.Provider>
)

export const HideOrgContent = () => (
  <NotificationContext.Provider value={mockNotifications}>
    <SidebarContext.Provider
      value={{ isSidebarOpen: true, closeSidebar: noop, openSidebar: noop, toggleSidebar: noop }}
    >
      <MainSidebar versions={mockVersions} hideOrgContent />
    </SidebarContext.Provider>
  </NotificationContext.Provider>
)
