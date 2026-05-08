export default {
  title: 'Navigation/SubNav',
}

import { PageSidebarContext } from '@/providers/page-sidebar-provider'
import { SubNav } from './SubNav'
import type { TNavItem } from '@/types'

const links: TNavItem[] = [
  { path: 'overview', text: 'Overview', iconVariant: 'HouseSimple' as const },
  { path: 'components', text: 'Components', iconVariant: 'Stack' as const },
  { path: 'deploys', text: 'Deploys', iconVariant: 'ShippingContainer' as const },
  { path: 'actions', text: 'Actions', iconVariant: 'SneakerMove' as const },
]

const linksWithSections: TNavItem[] = [
  { type: 'section', label: 'Overview' },
  { path: '/', text: 'Overview', iconVariant: 'HouseSimple' as const },
  { type: 'section', label: 'App' },
  { path: '/stacks', text: 'Stacks', iconVariant: 'Stack' as const },
  { path: '/sandbox', text: 'Sandbox', iconVariant: 'ShippingContainer' as const },
  { path: '/components', text: 'Components', iconVariant: 'Cards' as const },
  { path: '/roles', text: 'Roles', iconVariant: 'FileLock' as const },
  { path: '/policies', text: 'Policy reports', iconVariant: 'ShieldCheck' as const },
  { type: 'section', label: 'Customer' },
  { path: '/drift', text: 'Drift evaluation', iconVariant: 'ScanIcon' as const },
  { path: '/actions', text: 'Actions', iconVariant: 'TerminalWindow' as const },
  { path: '/workflows', text: 'Workflows', iconVariant: 'TreeStructure' as const },
  { path: '/runner', text: 'Install runner', iconVariant: 'SneakerMove' as const },
]

const mockPageSidebar = {
  isPageSidebarOpen: true,
  closePageSidebar: () => {},
  openPageSidebar: () => {},
  togglePageSidebar: () => {},
}

const mockPageSidebarCollapsed = {
  ...mockPageSidebar,
  isPageSidebarOpen: false,
}

export const Default = () => (
  <PageSidebarContext.Provider value={mockPageSidebar}>
    <div className="h-screen flex">
      <SubNav basePath="/org-123/installs/install-456" links={links} />
      <div className="p-8">Page content</div>
    </div>
  </PageSidebarContext.Provider>
)

export const WithSections = () => (
  <PageSidebarContext.Provider value={mockPageSidebar}>
    <div className="h-screen flex">
      <SubNav basePath="/org-123/installs/install-456" links={linksWithSections} />
      <div className="p-8">Page content</div>
    </div>
  </PageSidebarContext.Provider>
)

export const WithSectionsCollapsed = () => (
  <PageSidebarContext.Provider value={mockPageSidebarCollapsed}>
    <div className="h-screen flex">
      <SubNav basePath="/org-123/installs/install-456" links={linksWithSections} />
      <div className="p-8">Page content</div>
    </div>
  </PageSidebarContext.Provider>
)

export const MinimalLinks = () => (
  <PageSidebarContext.Provider value={mockPageSidebar}>
    <div className="h-screen flex">
      <SubNav
        basePath="/org-123/installs/install-456"
        links={[
          { path: 'overview', text: 'Overview', iconVariant: 'HouseSimple' as const },
          { path: 'settings', text: 'Settings', iconVariant: 'Gear' as const },
        ]}
      />
      <div className="p-8">Page content</div>
    </div>
  </PageSidebarContext.Provider>
)
