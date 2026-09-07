import type { ReactNode, RefObject } from 'react'
import { cn } from '@/utils/classnames'
import { Button } from '../atoms/Button'
import { Card } from '../atoms/Card'
import { Icon } from '../atoms/Icon'
import { Kbd } from '../atoms/Kbd'
import { Text } from '../atoms/Text'
import { useDashboardShell } from '../../providers/dashboard-shell-provider'

export interface IDashboardHeader {
  leading?: ReactNode
  actions?: ReactNode
  userMenu?: ReactNode
  mobileTriggerRef?: RefObject<HTMLSpanElement | null>
  scrolled?: boolean
  className?: string
}

const SidebarShortcut = ({ label }: { label: string }) => (
  <span className="flex items-center gap-2">
    <Text variant="caption">{label}</Text>
    <span className="inline-flex items-center gap-1">
      <Kbd>Alt</Kbd>
      <Kbd>S</Kbd>
    </span>
  </span>
)

export const DashboardHeader = ({
  leading,
  actions,
  userMenu,
  mobileTriggerRef,
  scrolled = false,
  className,
}: IDashboardHeader) => {
  const { desktopSidebarExpanded, openMobileSidebar, toggleSidebar } =
    useDashboardShell()

  const desktopLabel = desktopSidebarExpanded
    ? 'Collapse sidebar'
    : 'Expand sidebar'

  return (
    <Card
      as="header"
      padding="none"
      blur={scrolled ? 'lg' : 'none'}
      shadow={scrolled ? 'floating' : 'none'}
      className={cn(
        'sticky top-3 z-30 flex h-14 shrink-0 items-center gap-2 px-4 transition-[background-color,border-color,box-shadow,backdrop-filter]',
        !scrolled && 'border-transparent !bg-transparent backdrop-blur-none',
        className
      )}
    >
      <span ref={mobileTriggerRef} className="inline-flex md:hidden">
        <Button
          variant="ghost"
          iconOnly
          aria-label="Open navigation"
          aria-keyshortcuts="Alt+S"
          tooltip={<SidebarShortcut label="Open navigation" />}
          tooltipSide="bottom"
          onClick={openMobileSidebar}
        >
          <Icon variant="SidebarSimpleIcon" size={20} />
        </Button>
      </span>
      <span className="hidden md:inline-flex">
        <Button
          variant="ghost"
          iconOnly
          aria-label={desktopLabel}
          aria-keyshortcuts="Alt+S"
          tooltip={<SidebarShortcut label={desktopLabel} />}
          tooltipSide="bottom"
          onClick={toggleSidebar}
        >
          <Icon variant="SidebarSimpleIcon" size={20} />
        </Button>
      </span>
      <div className="flex min-w-0 flex-1 items-center">{leading}</div>
      {actions ? (
        <div className="flex shrink-0 items-center gap-2">{actions}</div>
      ) : null}
      {userMenu ? (
        <div className="hidden shrink-0 md:block">{userMenu}</div>
      ) : null}
    </Card>
  )
}
