'use client'

import {
  createContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react'
import { setSidebarCookie } from '@/actions/layout/main-sidebar-cookie'

interface ISidebarContext {
  isSidebarOpen?: boolean
  closeSidebar?: () => void
  openSidebar?: () => void
  toggleSidebar?: () => void
}

export const SidebarContext = createContext<ISidebarContext>({})

export const SidebarProvider = ({
  children,
  initIsSidebarOpen = false,
}: {
  children: ReactNode
  initIsSidebarOpen?: boolean
}) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(initIsSidebarOpen)

  const closeSidebar = useCallback(() => {
    setSidebarCookie(false)
    setIsSidebarOpen(false)
  }, [])

  const openSidebar = useCallback(() => {
    setSidebarCookie(true)
    setIsSidebarOpen(true)
  }, [])

  const toggleSidebar = useCallback(() => {
    setSidebarCookie(!isSidebarOpen)
    setIsSidebarOpen((prev) => !prev)
  }, [isSidebarOpen])

  // Add keyboard shortcut for Alt+S to toggle sidebar
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      // Alt+S (no Ctrl/Shift/Meta)
      if (
        e.altKey &&
        !e.shiftKey &&
        !e.ctrlKey &&
        !e.metaKey &&
        (e.key === 's' || e.key === 'S' || e.code === 'KeyS')
      ) {
        e.preventDefault()
        toggleSidebar()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [toggleSidebar])

  return (
    <SidebarContext.Provider
      value={{
        closeSidebar,
        isSidebarOpen,
        openSidebar,
        toggleSidebar,
      }}
    >
      {children}
    </SidebarContext.Provider>
  )
}
