export default {
  title: 'Playground/App/LandingPage',
  fullBleed: true,
}

import type { ReactNode } from 'react'
import { manyBranchCards } from '@/components/branches/BranchCards/BranchCards.fixtures'
import { mockAppInstalls } from '@/components/installs/AppInstallsList/AppInstallsList.fixtures'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { AppLandingPage } from './AppLandingPage'

const mockBreadcrumb = {
  breadcrumbLinks: [
    { path: '/org-1', text: 'Acme' },
    { path: '/org-1/apps', text: 'Apps' },
    { path: '/org-1/apps/app-1', text: 'acme-platform' },
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

const mockApp = {
  id: 'app-1',
  name: 'acme-platform',
  repo: 'acme/platform-configs',
  repoHref: 'https://github.com/acme/platform-configs',
}

export const Default = () => (
  <Providers>
    <div className="flex h-full min-h-0">
      <AppLandingPage
        app={mockApp}
        branches={manyBranchCards}
        installs={mockAppInstalls}
      />
    </div>
  </Providers>
)
