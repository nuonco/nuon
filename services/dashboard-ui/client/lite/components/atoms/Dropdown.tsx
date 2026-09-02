import {
  cloneElement,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
  type MouseEvent,
  type ReactElement,
  type ReactNode,
} from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/utils/classnames'
import {
  usePopover,
  type TPopoverAlign,
  type TPopoverSide,
} from '../../hooks/use-popover'
import { Tooltip, type ITooltip } from './Tooltip'

export type TDropdownHaspopup = 'menu' | 'dialog' | 'listbox' | 'grid'

const OPEN_KEY: Record<TPopoverSide, string> = {
  top: 'ArrowUp',
  bottom: 'ArrowDown',
  left: 'ArrowLeft',
  right: 'ArrowRight',
}

const OPPOSITE_SIDE: Record<TPopoverSide, TPopoverSide> = {
  top: 'bottom',
  bottom: 'top',
  left: 'right',
  right: 'left',
}

const CLOSE_KEY: Partial<Record<TPopoverSide, string>> = {
  left: 'ArrowRight',
  right: 'ArrowLeft',
}

const isVerticalSide = (side: TPopoverSide) => side === 'top' || side === 'bottom'

type TFocusEntry = (() => void) | null

interface IDropdownContext {
  close: () => void
  registerFocusFirst: (focus: TFocusEntry) => void
  registerFocusLast: (focus: TFocusEntry) => void
  registerSurface: (surface: HTMLElement) => void
  unregisterSurface: (surface: HTMLElement) => void
}

const DropdownContext = createContext<IDropdownContext | null>(null)

export const useDropdown = () => useContext(DropdownContext)

export interface IDropdown {
  trigger: ReactElement
  triggerTooltip?: Omit<ITooltip, 'children'>
  open?: boolean
  defaultOpen?: boolean
  onOpenChange?: (open: boolean) => void
  side?: TPopoverSide
  align?: TPopoverAlign
  haspopup?: TDropdownHaspopup
  matchTriggerWidth?: boolean
  stretch?: boolean
  contentClassName?: string
  className?: string
  children: ReactNode
}

type TTriggerProps = {
  id: string
  onClick?: (event: MouseEvent<HTMLElement>) => void
  onKeyDown?: (event: KeyboardEvent<HTMLElement>) => void
  'aria-haspopup': TDropdownHaspopup
  'aria-expanded': boolean
  'aria-controls': string
}

export const Dropdown = ({
  trigger,
  triggerTooltip,
  open: controlledOpen,
  defaultOpen = false,
  onOpenChange,
  side = 'bottom',
  align = 'start',
  haspopup = 'menu',
  matchTriggerWidth = false,
  stretch = false,
  contentClassName,
  className,
  children,
}: IDropdown) => {
  const parent = useContext(DropdownContext)
  const [uncontrolledOpen, setUncontrolledOpen] = useState(defaultOpen)
  const isControlled = controlledOpen !== undefined
  const isOpen = isControlled ? controlledOpen : uncontrolledOpen

  const [mounted, setMounted] = useState(defaultOpen)
  const [triggerWidth, setTriggerWidth] = useState<number>()

  const baseId = useId()
  const triggerId = `${baseId}-trigger`
  const contentId = `${baseId}-content`

  const focusFirst = useRef<TFocusEntry>(null)
  const focusLast = useRef<TFocusEntry>(null)
  const pendingFocus = useRef<'first' | 'last' | null>(null)
  const descendants = useRef(new Set<HTMLElement>())

  const {
    triggerRef,
    contentRef,
    side: placedSide,
    style,
  } = usePopover<HTMLSpanElement, HTMLDivElement>({ open: isOpen, side, align })

  const setOpen = useCallback(
    (next: boolean) => {
      if (next) setMounted(true)
      if (!isControlled) setUncontrolledOpen(next)
      onOpenChange?.(next)
    },
    [isControlled, onOpenChange]
  )

  const triggerButton = useCallback(
    () => triggerRef.current?.querySelector<HTMLElement>('button, a, [tabindex]'),
    [triggerRef]
  )

  const close = useCallback(
    ({ restoreFocus = true }: { restoreFocus?: boolean } = {}) => {
      pendingFocus.current = null
      const wasInside = contentRef.current?.contains(document.activeElement)
      setOpen(false)
      if (restoreFocus && wasInside) triggerButton()?.focus()
    },
    [contentRef, setOpen, triggerButton]
  )

  const openWith = useCallback(
    (focus: 'first' | 'last' | null) => {
      if (isOpen && focus) {
        const entry = focus === 'first' ? focusFirst.current : focusLast.current
        if (entry) entry()
        else contentRef.current?.focus()
        return
      }
      pendingFocus.current = focus
      setOpen(true)
    },
    [isOpen, contentRef, setOpen]
  )

  const context = useMemo<IDropdownContext>(
    () => ({
      close: () => {
        close()
        parent?.close()
      },
      registerFocusFirst: (focus) => {
        focusFirst.current = focus
      },
      registerFocusLast: (focus) => {
        focusLast.current = focus
      },
      registerSurface: (surface) => {
        descendants.current.add(surface)
        parent?.registerSurface(surface)
      },
      unregisterSurface: (surface) => {
        descendants.current.delete(surface)
        parent?.unregisterSurface(surface)
      },
    }),
    [close, parent]
  )

  useEffect(() => {
    const surface = contentRef.current
    if (!parent || !surface || !mounted) return
    parent.registerSurface(surface)
    return () => parent.unregisterSurface(surface)
  }, [parent, mounted, contentRef])

  useLayoutEffect(() => {
    if (!isOpen || !matchTriggerWidth) return
    setTriggerWidth(triggerRef.current?.offsetWidth)
  }, [isOpen, matchTriggerWidth, triggerRef])

  useEffect(() => {
    if (!isOpen || !mounted) return
    const intent = pendingFocus.current
    pendingFocus.current = null
    if (!intent) return

    const entry = intent === 'first' ? focusFirst.current : focusLast.current
    if (entry) entry()
    else contentRef.current?.focus()
  }, [isOpen, mounted, contentRef])

  useEffect(() => {
    if (!isOpen) return

    const onPointerDown = (event: PointerEvent) => {
      const target = event.target as Node
      if (triggerRef.current?.contains(target)) return
      if (contentRef.current?.contains(target)) return
      for (const surface of descendants.current) {
        if (surface.contains(target)) return
      }
      close({ restoreFocus: false })
    }

    document.addEventListener('pointerdown', onPointerDown)
    return () => document.removeEventListener('pointerdown', onPointerDown)
  }, [isOpen, close, triggerRef, contentRef])

  const onTriggerKeyDown = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key === OPEN_KEY[side]) {
      event.preventDefault()
      event.stopPropagation()
      openWith('first')
      return
    }
    if (isVerticalSide(side) && event.key === OPEN_KEY[OPPOSITE_SIDE[side]]) {
      event.preventDefault()
      event.stopPropagation()
      openWith('last')
      return
    }
    if (event.key === 'Escape' && isOpen) {
      event.preventDefault()
      close()
    }
  }

  const onTriggerClick = (event: MouseEvent<HTMLElement>) => {
    if (isOpen) {
      close()
      return
    }
    openWith(event.detail === 0 ? 'first' : null)
  }

  const triggerProps: TTriggerProps = {
    id: triggerId,
    'aria-haspopup': haspopup,
    'aria-expanded': isOpen,
    'aria-controls': contentId,
    onClick: (event) => {
      trigger.props.onClick?.(event)
      onTriggerClick(event)
    },
    onKeyDown: (event) => {
      trigger.props.onKeyDown?.(event)
      if (!event.defaultPrevented) onTriggerKeyDown(event)
    },
  }

  const anchor = (
    <span
      ref={triggerRef}
      className={cn('inline-flex', stretch ? 'w-full' : 'w-fit', className)}
    >
      {cloneElement(trigger, triggerProps)}
    </span>
  )

  return (
    <DropdownContext.Provider value={context}>
      {triggerTooltip ? (
        <Tooltip
          {...triggerTooltip}
          open={isOpen ? false : triggerTooltip.open}
        >
          {anchor}
        </Tooltip>
      ) : (
        anchor
      )}
      {mounted &&
        createPortal(
          <div
            ref={contentRef}
            id={contentId}
            tabIndex={-1}
            aria-labelledby={triggerId}
            data-state={isOpen ? 'open' : 'closed'}
            data-side={placedSide}
            style={{ ...style, minWidth: triggerWidth }}
            className={cn(
              'popover max-h-[var(--available)] overflow-y-auto rounded-lg outline-none',
              contentClassName
            )}
            onKeyDown={(event) => {
              if (event.key === 'Escape') {
                event.preventDefault()
                event.stopPropagation()
                close()
                return
              }
              if (event.key === 'Tab') {
                close({ restoreFocus: false })
                return
              }
              if (event.key === CLOSE_KEY[placedSide]) {
                event.preventDefault()
                event.stopPropagation()
                close()
              }
            }}
          >
            {children}
          </div>,
          document.body
        )}
    </DropdownContext.Provider>
  )
}
