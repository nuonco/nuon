export default {
  title: 'UI / Layout / Page layout',
}

import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
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

const Providers = ({
  children,
  breadcrumb = mockBreadcrumb,
}: {
  children: ReactNode
  breadcrumb?: typeof mockBreadcrumb
}) => (
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

export const BottomActions = () => (
  <Providers>
    <div className="flex h-screen w-full">
      <PageLayout>
        <PageContent>
          <PageSection>
            <div className="h-[90rem] rounded-md bg-cool-grey-100 dark:bg-dark-grey-800" />
            <div className="flex justify-end gap-3">
              <Button variant="secondary">View details</Button>
              <Button>Save changes</Button>
            </div>
            <div className="h-12" />
          </PageSection>
        </PageContent>
      </PageLayout>
    </div>
  </Providers>
)
