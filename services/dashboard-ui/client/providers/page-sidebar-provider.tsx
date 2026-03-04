import { createContext, useEffect, useState, type ReactNode } from 'react'
import { getPageSidebarOpen, setPageSidebarOpen } from '@/lib/cookies'

interface IPageSidebarContext {
  isPageSidebarOpen?: boolean
  closePageSidebar?: () => void
  openPageSidebar?: () => void
  togglePageSidebar?: () => void
}

export const PageSidebarContext = createContext<IPageSidebarContext>({})

export const PageSidebarProvider = ({
  children,
}: {
  children: ReactNode
}) => {
  const [isPageSidebarOpen, setIsPageSidebarOpen] = useState(
    () => getPageSidebarOpen()
  )

  function closePageSidebar() {
    setPageSidebarOpen(false)
    setIsPageSidebarOpen(false)
  }

  function openPageSidebar() {
    setPageSidebarOpen(true)
    setIsPageSidebarOpen(true)
  }

  function togglePageSidebar() {
    setPageSidebarOpen(!isPageSidebarOpen)
    setIsPageSidebarOpen((prev) => !prev)
  }

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
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
