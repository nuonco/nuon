import { createContext, useContext } from 'react'

export interface IShellContext {
  isSidebarOpen: boolean
  toggleSidebar: () => void
  showText: boolean
  toggleText: () => void
}

export const ShellContext = createContext<IShellContext>({
  isSidebarOpen: true,
  toggleSidebar: () => {},
  showText: true,
  toggleText: () => {},
})

export const useShell = () => useContext(ShellContext)
