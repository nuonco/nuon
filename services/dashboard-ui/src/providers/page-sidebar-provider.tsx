'use client'

import { createContext, useEffect, useState, type ReactNode } from 'react'
import { setPageSidebarCookie } from '@/actions/layout/page-sidebar-cookie'

interface IPageSidebarContext {
  isPageSidebarOpen?: boolean
  closePageSidebar?: () => void
  openPageSidebar?: () => void
  togglePageSidebar?: () => void
}

export const PageSidebarContext = createContext<IPageSidebarContext>({})

export const PageSidebarProvider = ({
  children,
  initIsPageSidebarOpen = true,
}: {
  children: ReactNode
  initIsPageSidebarOpen?: boolean
}) => {
  const [isPageSidebarOpen, setIsPageSidebarOpen] = useState(
    initIsPageSidebarOpen
  )

  function closePageSidebar() {
    setPageSidebarCookie(false)
    setIsPageSidebarOpen(false)
  }

  function openPageSidebar() {
    setPageSidebarCookie(true)
    setIsPageSidebarOpen(true)
  }

  function togglePageSidebar() {
    setPageSidebarCookie(!isPageSidebarOpen)
    setIsPageSidebarOpen((prev) => !prev)
  }

  // Add keyboard shortcut for Alt/Option+Shift+S to toggle sidebar (cross-platform)
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      // Cross-platform Alt/Option + Shift + S (no Ctrl/Meta)
      // e.altKey works for Alt (Win/Linux) and Option (Mac)
      // e.metaKey = Command (Mac)
      if (
        e.altKey &&
        e.shiftKey &&
        !e.ctrlKey &&
        !e.metaKey &&
        (e.key === 's' || e.key === 'S' || e.code === 'KeyS')
      ) {
        e.preventDefault()
        togglePageSidebar()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [togglePageSidebar])

  return (
    <PageSidebarContext.Provider
      value={{
        isPageSidebarOpen,
        closePageSidebar,
        openPageSidebar,
        togglePageSidebar,
      }}
    >
      {children}
    </PageSidebarContext.Provider>
  )
}
