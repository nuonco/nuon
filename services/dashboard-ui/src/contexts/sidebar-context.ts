import { createContext } from 'react'

interface ISidebarContext {
  isSidebarOpen?: boolean
  closeSidebar?: () => void
  openSidebar?: () => void
  toggleSidebar?: () => void
}

export const SidebarContext = createContext<ISidebarContext>({})
