import type { ReactNode, RefObject } from 'react'
import { cn } from '@/utils/classnames'
import { Button } from '../atoms/Button'
import { Card } from '../atoms/Card'
import { Icon } from '../atoms/Icon'
import { Link } from '../atoms/Link'
import { Logo } from '../atoms/Logo'
import type { INavItem } from '../molecules/NavLink'
import { useDashboardShell } from '../../providers/dashboard-shell-provider'
import { DashboardNav } from './DashboardNav'

export interface IDashboardSidebar {
  primaryNav: INavItem[]
  secondaryNav?: INavItem[]
  userMenu?: ReactNode
  containerRef?: RefObject<HTMLDivElement | null>
}

export const DashboardSidebar = ({
  primaryNav,
  secondaryNav = [],
  userMenu,
  containerRef,
}: IDashboardSidebar) => {
  const {
    desktop,
    desktopSidebarExpanded,
    mobileSidebarOpen,
    closeMobileSidebar,
  } = useDashboardShell()
  const collapsed = desktop && !desktopSidebarExpanded

  return (
    <div
      ref={containerRef}
      role={!desktop && mobileSidebarOpen ? 'dialog' : undefined}
      aria-modal={!desktop && mobileSidebarOpen ? true : undefined}
      aria-label={!desktop && mobileSidebarOpen ? 'Main navigation' : undefined}
      aria-hidden={!desktop && !mobileSidebarOpen ? true : undefined}
      tabIndex={!desktop ? -1 : undefined}
      className={cn(
        'fixed inset-y-3 left-3 z-50 w-[min(17.5rem,calc(100vw-1.5rem))] transition-transform duration-[250ms] ease-[cubic-bezier(0.65,0,0.35,1)] motion-reduce:transition-none',
        mobileSidebarOpen ? 'translate-x-0' : '-translate-x-[calc(100%+1rem)]',
        'md:static md:inset-auto md:z-auto md:my-3 md:ml-3 md:h-auto md:translate-x-0 md:transition-[width]',
        desktopSidebarExpanded ? 'md:w-56' : 'md:w-14'
      )}
    >
      <Card
        as="aside"
        padding="none"
        blur="lg"
        shadow="floating"
        className="flex size-full min-h-0 flex-col overflow-hidden"
      >
        <div className="flex h-14 shrink-0 items-center gap-2 p-3">
          <span
            className={cn(
              'block min-w-0 shrink-0 overflow-hidden transition-[width,transform] duration-[250ms] ease-[cubic-bezier(0.65,0,0.35,1)] motion-reduce:transition-none',
              collapsed ? 'w-4 translate-x-2' : 'w-[3.375rem] translate-x-0'
            )}
          >
            <Link
              href="/"
              aria-label="Nuon dashboard"
              className="block w-[3.375rem]"
              style={{ color: 'var(--text-primary)' }}
            >
              <Logo variant="wordmark" tone="mono" size={22} />
            </Link>
          </span>
          <span className="ml-auto inline-flex md:hidden">
            <Button
              variant="ghost"
              iconOnly
              aria-label="Close navigation"
              onClick={closeMobileSidebar}
            >
              <Icon variant="XIcon" size={18} />
            </Button>
          </span>
        </div>

        <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-2">
          <DashboardNav
            items={primaryNav}
            label="Main"
            collapsed={collapsed}
            onNavigate={closeMobileSidebar}
          />
          {secondaryNav.length ? (
            <DashboardNav
              items={secondaryNav}
              label="Resources"
              collapsed={collapsed}
              onNavigate={closeMobileSidebar}
              className="mt-auto"
            />
          ) : null}
          {!desktop && userMenu ? (
            <div className="mt-auto shrink-0 pt-2">{userMenu}</div>
          ) : null}
        </div>
      </Card>
    </div>
  )
}
