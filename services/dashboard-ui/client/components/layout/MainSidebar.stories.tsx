export default {
  title: 'Layout/MainSidebar',
}

import { SidebarContext } from '@/providers/sidebar-provider'
import { MainSidebar } from './MainSidebar'

const mockVersions = {
  api: { git_ref: 'abc1234', version: '1.2.3' },
  ui: { version: '4.5.6' },
}

const noop = () => {}

export const Open = () => (
  <SidebarContext.Provider
    value={{ isSidebarOpen: true, closeSidebar: noop, openSidebar: noop, toggleSidebar: noop }}
  >
    <MainSidebar versions={mockVersions} />
  </SidebarContext.Provider>
)

export const Collapsed = () => (
  <SidebarContext.Provider
    value={{ isSidebarOpen: false, closeSidebar: noop, openSidebar: noop, toggleSidebar: noop }}
  >
    <MainSidebar versions={mockVersions} />
  </SidebarContext.Provider>
)

export const HideOrgContent = () => (
  <SidebarContext.Provider
    value={{ isSidebarOpen: true, closeSidebar: noop, openSidebar: noop, toggleSidebar: noop }}
  >
    <MainSidebar versions={mockVersions} hideOrgContent />
  </SidebarContext.Provider>
)
