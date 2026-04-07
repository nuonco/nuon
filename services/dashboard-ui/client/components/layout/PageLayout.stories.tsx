export default {
  title: 'Layout/PageLayout',
}

import { BreadcrumbContext } from '@/providers/breadcrumb-provider'
import { SidebarContext } from '@/providers/sidebar-provider'
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

export const Default = () => (
  <BreadcrumbContext.Provider value={mockBreadcrumb}>
    <SidebarContext.Provider value={mockSidebar}>
      <PageLayout>
        <PageContent>
          <PageSection>
            <p>Page content goes here</p>
          </PageSection>
        </PageContent>
      </PageLayout>
    </SidebarContext.Provider>
  </BreadcrumbContext.Provider>
)

export const LoadingBreadcrumbs = () => (
  <BreadcrumbContext.Provider value={{ ...mockBreadcrumb, isLoading: true }}>
    <SidebarContext.Provider value={mockSidebar}>
      <PageLayout>
        <PageContent>
          <PageSection>
            <p>Loading breadcrumbs</p>
          </PageSection>
        </PageContent>
      </PageLayout>
    </SidebarContext.Provider>
  </BreadcrumbContext.Provider>
)

export const HideBreadcrumbs = () => (
  <BreadcrumbContext.Provider value={mockBreadcrumb}>
    <SidebarContext.Provider value={mockSidebar}>
      <PageLayout hideBreadcrumbs>
        <PageContent>
          <PageSection>
            <p>No breadcrumbs shown</p>
          </PageSection>
        </PageContent>
      </PageLayout>
    </SidebarContext.Provider>
  </BreadcrumbContext.Provider>
)

export const SinglePage = () => (
  <BreadcrumbContext.Provider value={mockBreadcrumb}>
    <SidebarContext.Provider value={mockSidebar}>
      <PageLayout variant="single-page">
        <PageContent>
          <PageSection>
            <p>Single-page layout with logo</p>
          </PageSection>
        </PageContent>
      </PageLayout>
    </SidebarContext.Provider>
  </BreadcrumbContext.Provider>
)
