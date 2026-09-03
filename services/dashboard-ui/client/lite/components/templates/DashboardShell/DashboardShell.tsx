import {
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type HTMLAttributes,
  type ReactNode,
  type UIEvent,
} from 'react'
import { useLocation } from 'react-router'
import { cn } from '@/utils/classnames'
import { useFocusContainment } from '../../../hooks/use-focus-containment'
import { useNavShortcuts } from '../../../hooks/use-nav-shortcuts'
import {
  DashboardShellProvider,
  useDashboardShell,
} from '../../../providers/dashboard-shell-provider'
import { ShellBackground } from '../../atoms/ShellBackground'
import type { INavItem } from '../../molecules/NavLink'
import { Card } from '../../atoms/Card'
import { DashboardHeader } from '../../organisms/DashboardHeader'
import { DashboardSidebar } from '../../organisms/DashboardSidebar'

export interface IDashboardShell
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  primaryNav: INavItem[]
  secondaryNav?: INavItem[]
  userMenu?: ReactNode
  headerLeading?: ReactNode
  headerActions?: ReactNode
  statusBar?: ReactNode
  children: ReactNode
  contentClassName?: string
  initialDesktopExpanded?: boolean
}

interface IDashboardShellLayout
  extends Omit<IDashboardShell, 'initialDesktopExpanded'> {}

const DashboardShellLayout = ({
  primaryNav,
  secondaryNav = [],
  userMenu,
  headerLeading,
  headerActions,
  statusBar,
  children,
  contentClassName,
  className,
  ...props
}: IDashboardShellLayout) => {
  const { desktop, mobileSidebarOpen, closeMobileSidebar } = useDashboardShell()
  const [contentScrolled, setContentScrolled] = useState(false)
  const location = useLocation()
  const previousPath = useRef(location.pathname)
  const previousMobileOpen = useRef(false)
  const sidebarRef = useRef<HTMLDivElement>(null)
  const mobileTriggerRef = useRef<HTMLSpanElement>(null)
  const mainColumnRef = useRef<HTMLDivElement>(null)
  const navigation = useMemo(
    () => [...primaryNav, ...secondaryNav],
    [primaryNav, secondaryNav]
  )

  useNavShortcuts(navigation)
  useFocusContainment({
    active: !desktop && mobileSidebarOpen,
    containerRef: sidebarRef,
    restoreFocusRef: mobileTriggerRef,
    onEscape: closeMobileSidebar,
  })

  useEffect(() => {
    if (previousPath.current === location.pathname) return
    previousPath.current = location.pathname
    closeMobileSidebar()
  }, [closeMobileSidebar, location.pathname])

  useLayoutEffect(() => {
    const main = mainColumnRef.current
    const sidebar = sidebarRef.current
    if (main) main.inert = !desktop && mobileSidebarOpen
    if (sidebar) sidebar.inert = !desktop && !mobileSidebarOpen
    if (!desktop && previousMobileOpen.current && !mobileSidebarOpen) {
      mobileTriggerRef.current?.querySelector<HTMLElement>('button')?.focus()
    }
    previousMobileOpen.current = mobileSidebarOpen
  }, [desktop, mobileSidebarOpen])

  const handleContentScroll = (event: UIEvent<HTMLElement>) => {
    setContentScrolled(event.currentTarget.scrollTop > 0)
  }

  return (
    <div
      className={cn(
        'relative isolate flex h-screen w-full flex-col overflow-hidden bg-surface-default',
        className
      )}
      {...props}
    >
      <ShellBackground />
      <div className="relative z-10 flex min-h-0 flex-1">
        <DashboardSidebar
          primaryNav={primaryNav}
          secondaryNav={secondaryNav}
          userMenu={!desktop ? userMenu : undefined}
          containerRef={sidebarRef}
        />

        <div
          data-dashboard-backdrop
          className={cn(
            'fixed inset-0 z-40 bg-black/40 transition-opacity duration-200 md:hidden',
            mobileSidebarOpen ? 'opacity-100' : 'pointer-events-none opacity-0'
          )}
          aria-hidden="true"
          onClick={closeMobileSidebar}
        />

        <div
          ref={mainColumnRef}
          className="flex min-w-0 flex-1 flex-col overflow-hidden"
        >
          <main
            data-dashboard-scroll
            className={cn(
              'flex min-h-0 flex-1 flex-col gap-8 overflow-y-auto px-4 pb-4',
              contentClassName
            )}
            onScroll={handleContentScroll}
          >
            <DashboardHeader
              leading={headerLeading}
              actions={headerActions}
              userMenu={desktop ? userMenu : undefined}
              mobileTriggerRef={mobileTriggerRef}
              scrolled={contentScrolled}
              className="mt-3"
            />
            {children}
          </main>
        </div>
      </div>
      {statusBar ? (
        <div className="relative z-10 shrink-0 px-4 pb-4">
          <Card
            as="footer"
            padding="none"
            blur="lg"
            shadow="floating"
            className="overflow-hidden"
          >
            {statusBar}
          </Card>
        </div>
      ) : null}
    </div>
  )
}

export const DashboardShell = ({
  initialDesktopExpanded,
  ...props
}: IDashboardShell) => (
  <DashboardShellProvider initialDesktopExpanded={initialDesktopExpanded}>
    <DashboardShellLayout {...props} />
  </DashboardShellProvider>
)
