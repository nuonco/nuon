export default {
  title: 'Layout/MainLayout',
}

import { SidebarContext } from '@/providers/sidebar-provider'
import { MainLayout } from './MainLayout'

const mockVersions = {
  api: { git_ref: 'abc1234', version: '1.2.3' },
  ui: { version: '4.5.6' },
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

export const SidebarOpen = () => (
  <SidebarContext.Provider value={mockSidebarOpen}>
    <MainLayout versions={mockVersions}>
      <div className="p-8">Page content</div>
    </MainLayout>
  </SidebarContext.Provider>
)

export const SidebarClosed = () => (
  <SidebarContext.Provider value={mockSidebarClosed}>
    <MainLayout versions={mockVersions}>
      <div className="p-8">Page content</div>
    </MainLayout>
  </SidebarContext.Provider>
)

export const HideOrgContent = () => (
  <SidebarContext.Provider value={mockSidebarOpen}>
    <MainLayout versions={mockVersions} hideOrgContent>
      <div className="p-8">Page content without org nav</div>
    </MainLayout>
  </SidebarContext.Provider>
)
