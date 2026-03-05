import {
  createContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react'
import { setSidebarOpen } from '@/lib/cookies'

interface ISidebarContext {
  isSidebarOpen?: boolean
  closeSidebar?: () => void
  openSidebar?: () => void
  toggleSidebar?: () => void
}

export const SidebarContext = createContext<ISidebarContext>({})

export const SidebarProvider = ({
  children,
  initIsSidebarOpen = true,
}: {
  children: ReactNode
  initIsSidebarOpen?: boolean
}) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(initIsSidebarOpen)

  const closeSidebar = useCallback(() => {
    setSidebarOpen(false)
    setIsSidebarOpen(false)
  }, [])

  const openSidebar = useCallback(() => {
    setSidebarOpen(true)
    setIsSidebarOpen(true)
  }, [])

  const toggleSidebar = useCallback(() => {
    setIsSidebarOpen((prev) => {
      setSidebarOpen(!prev)
      return !prev
    })
  }, [])

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
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
      value={{ closeSidebar, isSidebarOpen, openSidebar, toggleSidebar }}
    >
      {children}
    </SidebarContext.Provider>
  )
}
