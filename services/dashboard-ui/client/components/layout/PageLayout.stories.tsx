export default {
  title: 'Layout/PageLayout',
}

import type { ReactNode } from 'react'
import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
import { NotificationContext } from '@/providers/notification-provider'
import { PageLayout } from './PageLayout'
import { PageContent } from './PageContent'
import { PageSection } from './PageSection'

const mockBreadcrumb = {
  breadcrumbLinks: [
    { path: '/org-001', text: 'My org' },
    { path: '/org-001/installs', text: 'Installs' },
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

const Providers = ({ children, breadcrumb = mockBreadcrumb }: { children: ReactNode; breadcrumb?: typeof mockBreadcrumb }) => (
  <NotificationContext.Provider value={mockNotifications}>
    <BreadcrumbContext.Provider value={breadcrumb}>
      <SidebarContext.Provider value={mockSidebar}>
        {children}
      </SidebarContext.Provider>
    </BreadcrumbContext.Provider>
  </NotificationContext.Provider>
)

export const Default = () => (
  <Providers>
    <PageLayout>
      <PageContent>
        <PageSection>
          <p>Page content goes here</p>
        </PageSection>
      </PageContent>
    </PageLayout>
  </Providers>
)

export const LoadingBreadcrumbs = () => (
  <Providers breadcrumb={{ ...mockBreadcrumb, isLoading: true }}>
    <PageLayout>
      <PageContent>
        <PageSection>
          <p>Loading breadcrumbs</p>
        </PageSection>
      </PageContent>
    </PageLayout>
  </Providers>
)

export const HideBreadcrumbs = () => (
  <Providers>
    <PageLayout hideBreadcrumbs>
      <PageContent>
        <PageSection>
          <p>No breadcrumbs shown</p>
        </PageSection>
      </PageContent>
    </PageLayout>
  </Providers>
)

export const SinglePage = () => (
  <Providers>
    <PageLayout variant="single-page">
      <PageContent>
        <PageSection>
          <p>Single-page layout with logo</p>
        </PageSection>
      </PageContent>
    </PageLayout>
  </Providers>
)
