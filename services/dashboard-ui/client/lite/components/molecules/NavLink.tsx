import type { MouseEvent } from 'react'
import { NavLink as RouterNavLink } from 'react-router'
import { cn } from '@/utils/classnames'
import { Icon, type TIconVariant } from '../atoms/Icon'
import { Kbd } from '../atoms/Kbd'
import { Text } from '../atoms/Text'
import { Tooltip } from '../atoms/Tooltip'

export interface INavItem {
  href: string
  label: string
  icon: TIconVariant
  shortcut?: string
  external?: boolean
  end?: boolean
}

export interface INavLink extends INavItem {
  collapsed?: boolean
  onNavigate?: () => void
}

const LINK_CLASSES =
  'group grid h-9 w-full grid-cols-[1.5rem_minmax(0,1fr)] items-center overflow-hidden rounded-lg px-2 text-body text-secondary no-underline outline-none transition-colors ' +
  'hover:bg-surface-02 hover:text-primary active:bg-surface-03 ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring ' +
  'aria-[current=page]:bg-surface-accent aria-[current=page]:text-accent'

const Shortcut = ({ shortcut }: { shortcut: string }) => (
  <span className="inline-flex items-center gap-1">
    {shortcut
      .trim()
      .split(/\s+/)
      .map((key) => (
        <Kbd key={key}>{key.toUpperCase()}</Kbd>
      ))}
  </span>
)

export const NavLink = ({
  href,
  label,
  icon,
  shortcut,
  external = false,
  end,
  collapsed = false,
  onNavigate,
}: INavLink) => {
  const content = (
    <>
      <span className="flex size-6 shrink-0 items-center justify-center">
        <Icon variant={icon} size={18} weight="bold" />
      </span>
      <span
        aria-hidden={collapsed || undefined}
        className={cn(
          'flex min-w-0 items-center gap-2 overflow-hidden whitespace-nowrap pl-2 opacity-100 transition-opacity duration-[250ms] ease-[cubic-bezier(0.65,0,0.35,1)] motion-reduce:transition-none',
          collapsed && 'opacity-0'
        )}
      >
        <Text
          variant="caption"
          color="inherit"
          className="min-w-0 flex-1 truncate"
        >
          {label}
        </Text>
        {shortcut ? <Shortcut shortcut={shortcut} /> : null}
        {external ? (
          <Icon variant="ArrowSquareOutIcon" size={13} className="shrink-0" />
        ) : null}
      </span>
    </>
  )

  const onClick = (event: MouseEvent<HTMLAnchorElement>) => {
    if (!external && !event.defaultPrevented) onNavigate?.()
  }

  const link = external ? (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      aria-label={collapsed ? label : undefined}
      className={LINK_CLASSES}
    >
      {content}
      <span className="sr-only">(opens in a new tab)</span>
    </a>
  ) : (
    <RouterNavLink
      to={href}
      end={end}
      onClick={onClick}
      aria-label={collapsed ? label : undefined}
      className={LINK_CLASSES}
    >
      {content}
    </RouterNavLink>
  )

  return (
    <Tooltip
      content={
        <span className="flex items-center gap-2">
          <Text variant="caption">{label}</Text>
          {shortcut ? <Shortcut shortcut={shortcut} /> : null}
          {external ? <Icon variant="ArrowSquareOutIcon" size={13} /> : null}
        </span>
      }
      side="right"
      disableHover={!collapsed}
      className="w-full"
    >
      {link}
    </Tooltip>
  )
}
