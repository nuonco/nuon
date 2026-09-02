import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  type HTMLAttributes,
  type KeyboardEvent,
  type ReactNode,
} from 'react'
import { Link as RouterLink } from 'react-router'
import { cn } from '@/utils/classnames'
import { Dropdown, useDropdown } from '../atoms/Dropdown'
import { Icon } from '../atoms/Icon'
import { isExternalHref } from '../atoms/Link'
import { Text } from '../atoms/Text'

const ITEM_SELECTOR = '[role^="menuitem"]:not([aria-disabled="true"])'

const TYPEAHEAD_RESET = 500

export interface IMenu extends Omit<HTMLAttributes<HTMLDivElement>, 'role'> {}

export const Menu = ({ className, children, ...props }: IMenu) => {
  const dropdown = useDropdown()
  const ref = useRef<HTMLDivElement>(null)
  const typeahead = useRef({ query: '', at: 0 })

  const items = useCallback(
    () => Array.from(ref.current?.querySelectorAll<HTMLElement>(ITEM_SELECTOR) ?? []),
    []
  )

  const syncTabStops = useCallback(() => {
    const list = items()
    const active = list.find((item) => item === document.activeElement) ?? list[0]
    for (const item of list) item.tabIndex = item === active ? 0 : -1
  }, [items])

  useLayoutEffect(syncTabStops, [children, syncTabStops])

  useEffect(() => {
    if (!dropdown) return
    dropdown.registerFocusFirst(() => items().at(0)?.focus())
    dropdown.registerFocusLast(() => items().at(-1)?.focus())
    return () => {
      dropdown.registerFocusFirst(null)
      dropdown.registerFocusLast(null)
    }
  }, [dropdown, items])

  const move = (offset: number, event: KeyboardEvent) => {
    event.preventDefault()
    event.stopPropagation()
    const list = items()
    if (!list.length) return
    const current = list.indexOf(document.activeElement as HTMLElement)
    const next =
      current === -1
        ? offset > 0
          ? 0
          : list.length - 1
        : (current + offset + list.length) % list.length
    list[next]?.focus()
  }

  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === 'ArrowDown') return move(1, event)
    if (event.key === 'ArrowUp') return move(-1, event)
    if (event.key === 'Home') {
      event.preventDefault()
      event.stopPropagation()
      return items().at(0)?.focus()
    }
    if (event.key === 'End') {
      event.preventDefault()
      event.stopPropagation()
      return items().at(-1)?.focus()
    }
    if (event.key.length !== 1 || event.metaKey || event.ctrlKey || event.altKey) {
      return
    }

    const now = performance.now()
    typeahead.current = {
      query:
        now - typeahead.current.at > TYPEAHEAD_RESET
          ? event.key.toLowerCase()
          : typeahead.current.query + event.key.toLowerCase(),
      at: now,
    }

    const match = items().find((item) =>
      item.textContent?.trim().toLowerCase().startsWith(typeahead.current.query)
    )
    if (match) {
      event.preventDefault()
      event.stopPropagation()
      match.focus()
    }
  }

  return (
    <div
      ref={ref}
      role="menu"
      className={cn('flex min-w-48 flex-col gap-0.5 p-1', className)}
      onKeyDown={onKeyDown}
      onFocusCapture={syncTabStops}
      {...props}
    >
      {children}
    </div>
  )
}

export type TMenuItemTone = 'default' | 'danger'

export interface IMenuItem {
  icon?: ReactNode
  href?: string
  external?: boolean
  selected?: boolean
  disabled?: boolean
  tone?: TMenuItemTone
  closeOnSelect?: boolean
  onSelect?: () => void
  className?: string
  children: ReactNode
}

const ITEM_CLASSES =
  'flex h-8 w-full shrink-0 cursor-pointer items-center gap-2 rounded-md border border-transparent px-2 ' +
  'text-left text-body no-underline outline-none transition-colors ' +
  'focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-focus-ring'

const TONE_CLASSES: Record<TMenuItemTone, string> = {
  default:
    'text-menu-item-text not-aria-disabled:hover:bg-menu-item-hover not-aria-disabled:focus:bg-menu-item-hover ' +
    'not-aria-disabled:active:bg-menu-item-active',
  danger:
    'text-menu-item-danger not-aria-disabled:hover:bg-menu-item-danger-hover ' +
    'not-aria-disabled:focus:bg-menu-item-danger-hover not-aria-disabled:active:bg-menu-item-danger-hover',
}

export const MenuItem = ({
  icon,
  href,
  external,
  selected,
  disabled = false,
  tone = 'default',
  closeOnSelect = true,
  onSelect,
  className,
  children,
}: IMenuItem) => {
  const dropdown = useDropdown()
  const isCheckable = selected !== undefined

  const select = () => {
    if (disabled) return
    onSelect?.()
    if (closeOnSelect) dropdown?.close()
  }

  const shared = {
    role: isCheckable ? 'menuitemcheckbox' : 'menuitem',
    tabIndex: -1,
    'aria-checked': isCheckable ? selected : undefined,
    'aria-disabled': disabled || undefined,
    className: cn(
      ITEM_CLASSES,
      TONE_CLASSES[tone],
      selected && 'bg-menu-item-selected',
      disabled && 'cursor-not-allowed opacity-50',
      className
    ),
  }

  const body = (
    <>
      {icon ? <span className="flex shrink-0 items-center">{icon}</span> : null}
      <span className="min-w-0 flex-1 truncate">{children}</span>
      {selected ? (
        <Icon variant="CheckIcon" size={16} className="shrink-0" aria-hidden />
      ) : null}
    </>
  )

  if (href && !disabled) {
    const isExternal = external ?? isExternalHref(href)

    if (isExternal) {
      return (
        <a
          {...shared}
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          onClick={select}
        >
          {body}
          <Icon
            variant="ArrowSquareOutIcon"
            size={14}
            className="shrink-0"
            aria-hidden
          />
          <span className="sr-only">(opens in a new tab)</span>
        </a>
      )
    }

    return (
      <RouterLink {...shared} to={href} onClick={select}>
        {body}
      </RouterLink>
    )
  }

  return (
    <button {...shared} type="button" onClick={select}>
      {body}
    </button>
  )
}

export interface IMenuSubmenu {
  icon?: ReactNode
  label: ReactNode
  disabled?: boolean
  className?: string
  children: ReactNode
}

export const MenuSubmenu = ({
  icon,
  label,
  disabled = false,
  className,
  children,
}: IMenuSubmenu) => (
  <Dropdown
    stretch
    side="right"
    align="start"
    trigger={
      <button
        type="button"
        role="menuitem"
        tabIndex={-1}
        aria-disabled={disabled || undefined}
        className={cn(
          ITEM_CLASSES,
          TONE_CLASSES.default,
          disabled && 'cursor-not-allowed opacity-50',
          className
        )}
      >
        {icon ? <span className="flex shrink-0 items-center">{icon}</span> : null}
        <span className="min-w-0 flex-1 truncate">{label}</span>
        <Icon variant="CaretRightIcon" size={16} className="shrink-0" aria-hidden />
      </button>
    }
  >
    {children}
  </Dropdown>
)

export interface IMenuLabel extends HTMLAttributes<HTMLDivElement> {}

export const MenuLabel = ({ className, children, ...props }: IMenuLabel) => (
  <Text
    as="div"
    variant="label"
    color="tertiary"
    className={cn('px-2 pt-2 pb-1', className)}
    {...props}
  >
    {children}
  </Text>
)

export interface IMenuSeparator extends HTMLAttributes<HTMLHRElement> {}

export const MenuSeparator = ({ className, ...props }: IMenuSeparator) => (
  <hr
    role="separator"
    className={cn('my-1 border-0 border-t border-divider', className)}
    {...props}
  />
)
