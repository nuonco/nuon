export default {
  title: 'Installs/InstallDetailLayoutPlayground',
  fullBleed: true,
}

import type { ReactNode } from 'react'
import { AuthContext } from '@/providers/auth-provider'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { DashboardPreferencesProvider } from '@/providers/dashboard-preferences-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { PageSidebarContext } from '@/providers/page-sidebar-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import {
  branchMovedFixture,
  configCurrentFixture,
  infraDriftFixture,
  resourceLagFixture,
} from '../InstallDetailPlayground/fixtures'
import { InstallDetailLayoutPlayground } from './InstallDetailLayoutPlayground'

const mockBreadcrumb = {
  breadcrumbLinks: [
    {
      path: `/${configCurrentFixture.orgId}`,
      text: configCurrentFixture.orgName,
    },
    { path: `/${configCurrentFixture.orgId}/installs`, text: 'Installs' },
    {
      path: `/${configCurrentFixture.orgId}/installs/${configCurrentFixture.id}`,
      text: configCurrentFixture.name,
    },
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

const mockPageSidebar = {
  isPageSidebarOpen: true,
  closePageSidebar: () => {},
  openPageSidebar: () => {},
  togglePageSidebar: () => {},
}

const mockAuth = {
  user: { sub: '1', email: 'test@nuon.co' },
  isAuthenticated: true,
  isAdmin: true,
  isNuonEmployee: true,
  isLoading: false,
  error: null,
  demoMode: false,
  toggleDemoMode: () => {},
}

const Providers = ({
  children,
  isPageSidebarOpen = true,
}: {
  children: ReactNode
  isPageSidebarOpen?: boolean
}) => (
  <AuthContext.Provider value={mockAuth}>
    <DashboardPreferencesProvider>
      <NotificationContext.Provider value={mockNotifications}>
        <BreadcrumbContext.Provider value={mockBreadcrumb}>
          <SidebarContext.Provider value={mockSidebar}>
            <PageSidebarContext.Provider
              value={{ ...mockPageSidebar, isPageSidebarOpen }}
            >
              {children}
            </PageSidebarContext.Provider>
          </SidebarContext.Provider>
        </BreadcrumbContext.Provider>
      </NotificationContext.Provider>
    </DashboardPreferencesProvider>
  </AuthContext.Provider>
)

export const ConfigCurrent = () => (
  <Providers>
    <InstallDetailLayoutPlayground install={configCurrentFixture} />
  </Providers>
)

ConfigCurrent.storyName = 'Config current'

export const BranchMoved = () => (
  <Providers>
    <InstallDetailLayoutPlayground install={branchMovedFixture} />
  </Providers>
)

BranchMoved.storyName = 'Branch moved'

export const ResourceLag = () => (
  <Providers>
    <InstallDetailLayoutPlayground install={resourceLagFixture} />
  </Providers>
)

ResourceLag.storyName = 'Resource lag'

export const InfraDrift = () => (
  <Providers>
    <InstallDetailLayoutPlayground install={infraDriftFixture} />
  </Providers>
)

InfraDrift.storyName = 'Infra drift'

export const CollapsedNav = () => (
  <Providers isPageSidebarOpen={false}>
    <InstallDetailLayoutPlayground install={configCurrentFixture} />
  </Providers>
)

CollapsedNav.storyName = 'Collapsed nav'
