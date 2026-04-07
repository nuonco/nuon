export default {
  title: 'Layout/MainTopbar',
}

import { SidebarProvider } from '@/providers/sidebar-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { MainTopbar } from './MainTopbar'
import { Text } from '@/components/common/Text'

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

export const Default = () => (
  <NotificationContext.Provider value={mockNotifications}>
    <SidebarProvider>
      <MainTopbar>
        <Text variant="subtext" theme="neutral">Dashboard</Text>
      </MainTopbar>
    </SidebarProvider>
  </NotificationContext.Provider>
)

export const HideSidebarButtons = () => (
  <NotificationContext.Provider value={mockNotifications}>
    <SidebarProvider>
      <MainTopbar hideSidebarButtons>
        <Text variant="subtext" theme="neutral">Single page layout</Text>
      </MainTopbar>
    </SidebarProvider>
  </NotificationContext.Provider>
)
