export default {
  title: 'Navigation/SubNavLink',
}

import { PageSidebarContext } from '@/providers/page-sidebar-provider'
import { SubNavLink } from './SubNavLink'

const mockPageSidebar = {
  isPageSidebarOpen: true,
  closePageSidebar: () => {},
  openPageSidebar: () => {},
  togglePageSidebar: () => {},
}

export const Default = () => (
  <PageSidebarContext.Provider value={mockPageSidebar}>
    <nav className="flex flex-col gap-1 p-4 w-[280px]">
      <SubNavLink
        basePath="/org-123/installs/install-456"
        path="overview"
        text="Overview"
        iconVariant="HouseSimpleIcon"
      />
      <SubNavLink
        basePath="/org-123/installs/install-456"
        path="components"
        text="Components"
        iconVariant="StackIcon"
      />
      <SubNavLink
        basePath="/org-123/installs/install-456"
        path="deploys"
        text="Deploys"
        iconVariant="ShippingContainerIcon"
      />
    </nav>
  </PageSidebarContext.Provider>
)

export const NoIcon = () => (
  <PageSidebarContext.Provider value={mockPageSidebar}>
    <nav className="flex flex-col gap-1 p-4 w-[280px]">
      <SubNavLink
        basePath="/org-123/installs/install-456"
        path="overview"
        text="Overview"
      />
      <SubNavLink
        basePath="/org-123/installs/install-456"
        path="settings"
        text="Settings"
      />
    </nav>
  </PageSidebarContext.Provider>
)
