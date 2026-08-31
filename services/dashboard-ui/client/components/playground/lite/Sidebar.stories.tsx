import { Sidebar } from './Sidebar'
import { ShellContext } from './shell-context'
import { primaryNav, secondaryNav } from './nav'

export default {
  title: 'Playground/Lite/Sidebar',
}

export const Default = () => (
  <div className="flex h-[600px]">
    <Sidebar primaryNav={primaryNav} secondaryNav={secondaryNav} />
  </div>
)

export const Collapsed = () => (
  <ShellContext.Provider
    value={{
      isSidebarOpen: false,
      toggleSidebar: () => {},
      showText: true,
      toggleText: () => {},
    }}
  >
    <div className="flex h-[600px]">
      <Sidebar primaryNav={primaryNav} secondaryNav={secondaryNav} />
    </div>
  </ShellContext.Provider>
)

export const PrimaryOnly = () => (
  <div className="flex h-[600px]">
    <Sidebar primaryNav={primaryNav} />
  </div>
)
