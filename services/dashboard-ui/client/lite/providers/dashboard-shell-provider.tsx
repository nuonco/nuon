import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { useMediaQuery } from '../hooks/use-media-query'

export const DASHBOARD_DESKTOP_QUERY = '(min-width: 768px)'
export const DASHBOARD_SIDEBAR_STORAGE_KEY =
  'nuon-lite:dashboard-sidebar-expanded'

const readDesktopPreference = () => {
  try {
    const value = localStorage.getItem(DASHBOARD_SIDEBAR_STORAGE_KEY)
    return value === null ? true : value === 'true'
  } catch {
    return true
  }
}

const writeDesktopPreference = (expanded: boolean) => {
  try {
    localStorage.setItem(DASHBOARD_SIDEBAR_STORAGE_KEY, String(expanded))
  } catch {
    return
  }
}

interface IDashboardShellContext {
  desktop: boolean
  desktopSidebarExpanded: boolean
  mobileSidebarOpen: boolean
  closeMobileSidebar: () => void
  collapseDesktopSidebar: () => void
  expandDesktopSidebar: () => void
  openMobileSidebar: () => void
  toggleSidebar: () => void
}

const DashboardShellContext = createContext<IDashboardShellContext | null>(null)

export interface IDashboardShellProvider {
  children: ReactNode
  initialDesktopExpanded?: boolean
}

export const DashboardShellProvider = ({
  children,
  initialDesktopExpanded,
}: IDashboardShellProvider) => {
  const desktop = useMediaQuery(DASHBOARD_DESKTOP_QUERY)
  const [desktopSidebarExpanded, setDesktopSidebarExpanded] = useState(
    () => initialDesktopExpanded ?? readDesktopPreference()
  )
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false)

  const setDesktopPreference = useCallback((expanded: boolean) => {
    setDesktopSidebarExpanded(expanded)
    writeDesktopPreference(expanded)
  }, [])

  const expandDesktopSidebar = useCallback(
    () => setDesktopPreference(true),
    [setDesktopPreference]
  )
  const collapseDesktopSidebar = useCallback(
    () => setDesktopPreference(false),
    [setDesktopPreference]
  )
  const openMobileSidebar = useCallback(() => setMobileSidebarOpen(true), [])
  const closeMobileSidebar = useCallback(() => setMobileSidebarOpen(false), [])
  const toggleSidebar = useCallback(() => {
    if (desktop) {
      setDesktopSidebarExpanded((expanded) => {
        writeDesktopPreference(!expanded)
        return !expanded
      })
      return
    }
    setMobileSidebarOpen((open) => !open)
  }, [desktop])

  useEffect(() => {
    if (desktop) closeMobileSidebar()
  }, [closeMobileSidebar, desktop])

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (
        !event.altKey ||
        event.shiftKey ||
        event.ctrlKey ||
        event.metaKey ||
        event.code !== 'KeyS'
      ) {
        return
      }
      event.preventDefault()
      toggleSidebar()
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [toggleSidebar])

  const value = useMemo(
    () => ({
      desktop,
      desktopSidebarExpanded,
      mobileSidebarOpen,
      closeMobileSidebar,
      collapseDesktopSidebar,
      expandDesktopSidebar,
      openMobileSidebar,
      toggleSidebar,
    }),
    [
      desktop,
      desktopSidebarExpanded,
      mobileSidebarOpen,
      closeMobileSidebar,
      collapseDesktopSidebar,
      expandDesktopSidebar,
      openMobileSidebar,
      toggleSidebar,
    ]
  )

  return (
    <DashboardShellContext.Provider value={value}>
      {children}
    </DashboardShellContext.Provider>
  )
}

export const useDashboardShell = () => {
  const context = useContext(DashboardShellContext)
  if (!context) {
    throw new Error(
      'useDashboardShell must be used within DashboardShellProvider'
    )
  }
  return context
}
