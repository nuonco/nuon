import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/utils/classnames'
import { Button, type IButtonAsButton } from './Button'
import { Icon } from './Icon'
import { TransitionDiv } from './TransitionDiv'
import './Dropdown.css'

type TDropdownNestingContext = {
  registerChild: (el: HTMLElement) => void
  unregisterChild: (el: HTMLElement) => void
}

const DropdownNestingContext =
  createContext<TDropdownNestingContext | null>(null)

const MENU_ITEM_SELECTOR =
  'button:not([data-focus-guard]), a, [role="menuitem"], [tabindex]:not([tabindex="-1"]):not([data-focus-guard])'

export interface IDropdown extends IButtonAsButton {
  alignment?: 'left' | 'right' | 'overlay'
  buttonClassName?: string
  buttonText: React.ReactNode
  children: React.ReactNode
  closeOnBlur?: boolean
  dropdownClassName?: string
  hideIcon?: boolean
  icon?: React.ReactNode
  iconAlignment?: 'left' | 'right'
  isOpen?: boolean
  onOpenChange?: (isOpen: boolean) => void
  id: string
  position?: 'above' | 'below' | 'beside' | 'overlay'
  wrapperClassName?: string
}

export const Dropdown = ({
  alignment = 'left',
  buttonText,
  buttonClassName,
  children,
  className,
  closeOnBlur = true,
  dropdownClassName,
  hideIcon = false,
  icon = <Icon variant="CaretDownIcon" />,
  iconAlignment = 'right',
  id,
  isOpen: initIsOpen = false,
  onOpenChange,
  position = 'below',
  variant,
  ...props
}: IDropdown) => {
  const [isOpen, setIsOpen] = useState(initIsOpen)
  const [styles, setStyles] = useState<React.CSSProperties>({})
  const [placement, setPlacement] = useState<{
    position: NonNullable<IDropdown['position']>
    alignment: NonNullable<IDropdown['alignment']>
  }>({ position, alignment })
  const [contentEl, setContentEl] = useState<HTMLDivElement | null>(null)
  const triggerRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement | null>(null)
  const childPortals = useRef<Set<HTMLElement>>(new Set())
  const pendingFocus = useRef(false)
  const parentNesting = useContext(DropdownNestingContext)

  const handleClose = useCallback(() => {
    pendingFocus.current = false
    setIsOpen(false)
  }, [])

  const hasMounted = useRef(false)
  useEffect(() => {
    if (!hasMounted.current) {
      hasMounted.current = true
      return
    }
    onOpenChange?.(isOpen)
  }, [isOpen, onOpenChange])

  const focusFirstItem = useCallback(() => {
    const item = contentRef.current?.querySelector<HTMLElement>(MENU_ITEM_SELECTOR)
    if (item) item.focus()
    else pendingFocus.current = true
  }, [])

  const nestingContext = useRef<TDropdownNestingContext>({
    registerChild: (el) => {
      childPortals.current.add(el)
      parentNesting?.registerChild(el)
    },
    unregisterChild: (el) => {
      childPortals.current.delete(el)
      parentNesting?.unregisterChild(el)
    },
  }).current

  const contentCallbackRef = useCallback(
    (el: HTMLDivElement | null) => {
      const prev = contentRef.current
      contentRef.current = el
      setContentEl(el)

      if (parentNesting) {
        if (prev) parentNesting.unregisterChild(prev)
        if (el) parentNesting.registerChild(el)
      }
    },
    [parentNesting]
  )

  useEffect(() => {
    if (!isOpen || !contentEl || !pendingFocus.current) return
    pendingFocus.current = false
    contentEl.querySelector<HTMLElement>(MENU_ITEM_SELECTOR)?.focus()
  }, [isOpen, contentEl])

  const isInsideTree = useCallback(
    (target: Node | null): boolean => {
      if (!target) return false
      if (triggerRef.current?.contains(target)) return true
      if (contentRef.current?.contains(target)) return true
      for (const child of childPortals.current) {
        if (child.contains(target)) return true
      }
      return false
    },
    []
  )

  const calculatePosition = useCallback(() => {
    if (!triggerRef.current) return

    const trigger = triggerRef.current.getBoundingClientRect()
    const content = contentRef.current
    const contentWidth = content?.offsetWidth ?? 0
    const contentHeight = content?.offsetHeight ?? 0
    const vw = window.innerWidth
    const vh = window.innerHeight
    const gap = 8
    const margin = 8

    let effPosition = position
    let effAlignment = alignment

    if (position === 'below' || position === 'above') {
      const spaceBelow = vh - trigger.bottom
      const spaceAbove = trigger.top
      const needed = contentHeight + gap
      if (position === 'below' && needed > spaceBelow && spaceAbove > spaceBelow) {
        effPosition = 'above'
      } else if (
        position === 'above' &&
        needed > spaceAbove &&
        spaceBelow > spaceAbove
      ) {
        effPosition = 'below'
      }
    }

    if (position === 'beside') {
      const spaceRight = vw - trigger.right
      const spaceLeft = trigger.left
      const needed = contentWidth + gap
      if (alignment === 'right' && needed > spaceRight && spaceLeft > spaceRight) {
        effAlignment = 'left'
      } else if (
        alignment === 'left' &&
        needed > spaceLeft &&
        spaceRight > spaceLeft
      ) {
        effAlignment = 'right'
      }
    } else if (contentWidth) {
      if (
        alignment === 'left' &&
        trigger.left + contentWidth > vw - margin &&
        trigger.right - contentWidth >= margin
      ) {
        effAlignment = 'right'
      } else if (
        alignment === 'right' &&
        trigger.right - contentWidth < margin &&
        trigger.left + contentWidth <= vw - margin
      ) {
        effAlignment = 'left'
      }
    }

    const newStyles: React.CSSProperties = {
      position: 'fixed',
      zIndex: 60,
    }

    if (effPosition === 'below') {
      newStyles.top = trigger.bottom + gap
      if (effAlignment === 'left') newStyles.left = trigger.left
      if (effAlignment === 'right') newStyles.right = vw - trigger.right
    } else if (effPosition === 'above') {
      newStyles.bottom = vh - trigger.top + gap
      if (effAlignment === 'left') newStyles.left = trigger.left
      if (effAlignment === 'right') newStyles.right = vw - trigger.right
    } else if (effPosition === 'beside') {
      if (effAlignment === 'left') newStyles.right = vw - trigger.left + gap
      if (effAlignment === 'right') newStyles.left = trigger.right + gap
      newStyles.top =
        contentHeight && trigger.top + contentHeight > vh - margin
          ? Math.max(margin, vh - contentHeight - margin)
          : trigger.top
    } else if (effPosition === 'overlay') {
      newStyles.top = trigger.top
      newStyles.left = trigger.left
    }

    setStyles(newStyles)
    setPlacement((prev) =>
      prev.position === effPosition && prev.alignment === effAlignment
        ? prev
        : { position: effPosition, alignment: effAlignment }
    )
  }, [position, alignment])

  useLayoutEffect(() => {
    if (!isOpen) return

    calculatePosition()

    const resizeObserver = contentEl
      ? new ResizeObserver(() => calculatePosition())
      : null
    if (contentEl && resizeObserver) resizeObserver.observe(contentEl)

    window.addEventListener('resize', calculatePosition)
    window.addEventListener('scroll', calculatePosition, true)
    return () => {
      resizeObserver?.disconnect()
      window.removeEventListener('resize', calculatePosition)
      window.removeEventListener('scroll', calculatePosition, true)
    }
  }, [isOpen, contentEl, calculatePosition])

  useEffect(() => {
    if (!isOpen) return

    const handleClickOutside = (event: MouseEvent) => {
      if (!isInsideTree(event.target as Node)) {
        handleClose()
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen, handleClose, isInsideTree])

  useEffect(() => {
    if (!isOpen || !closeOnBlur) return

    const triggerEl = triggerRef.current
    const contentEl = contentRef.current
    const handleFocusOut = (event: FocusEvent) => {
      if (!isInsideTree(event.relatedTarget as Node)) {
        handleClose()
      }
    }

    triggerEl?.addEventListener('focusout', handleFocusOut, true)
    contentEl?.addEventListener('focusout', handleFocusOut, true)
    return () => {
      triggerEl?.removeEventListener('focusout', handleFocusOut, true)
      contentEl?.removeEventListener('focusout', handleFocusOut, true)
    }
  }, [isOpen, closeOnBlur, handleClose, isInsideTree])

  const dropdownContent = (
    <TransitionDiv
      ref={contentCallbackRef}
      className={cn(
        'dropdown-content',
        'border',
        'rounded-lg',
        'shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08),0px_10px_32px_0px_rgba(0,0,0,0.08)]',
        'outline-none',
        'bg-white',
        'dark:bg-dark-grey-900',
        'w-fit',
        placement.alignment,
        placement.position,
        dropdownClassName
      )}
      aria-labelledby={`dropdown-button-${id}`}
      id={`dropdown-content-${id}`}
      isVisible={isOpen}
      style={styles}
      role="menu"
      tabIndex={-1}
      onKeyDown={(e) => {
        if (e.key === 'Escape') {
          e.preventDefault()
          handleClose()
          const trigger = triggerRef.current?.querySelector<HTMLElement>('button')
          trigger?.focus()
          return
        }
        if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
          e.preventDefault()
          const content = contentRef.current
          if (!content) return
          const items = Array.from(
            content.querySelectorAll<HTMLElement>(
              'button:not([data-focus-guard]), a, [role="menuitem"], [tabindex]:not([tabindex="-1"]):not([data-focus-guard])'
            )
          )
          if (!items.length) return
          const current = items.indexOf(document.activeElement as HTMLElement)
          let next: number
          if (e.key === 'ArrowDown') {
            next = current < items.length - 1 ? current + 1 : 0
          } else {
            next = current > 0 ? current - 1 : items.length - 1
          }
          items[next]?.focus()
        }
      }}
      onClick={(e) => {
        if (!closeOnBlur) return
        const target = e.target as HTMLElement
        if (target.closest('button, a, [role="menuitem"]')) {
          handleClose()
        }
      }}
    >
      <span
        tabIndex={0}
        data-focus-guard
        aria-hidden="true"
        style={{ position: 'fixed', opacity: 0, pointerEvents: 'none' }}
        onFocus={() => {
          handleClose()
          triggerRef.current?.querySelector<HTMLElement>('button')?.focus()
        }}
      />
      <DropdownNestingContext.Provider value={nestingContext}>
        {children}
      </DropdownNestingContext.Provider>
      <span
        tabIndex={0}
        data-focus-guard
        aria-hidden="true"
        style={{ position: 'fixed', opacity: 0, pointerEvents: 'none' }}
        onFocus={() => {
          handleClose()
          triggerRef.current?.querySelector<HTMLElement>('button')?.focus()
        }}
      />
    </TransitionDiv>
  )

  return (
    <div
      className={cn(
        'dropdown relative inline-block text-left leading-none',
        className
      )}
      id={id}
      ref={triggerRef}
    >
      <Button
        aria-haspopup="true"
        aria-expanded={isOpen}
        aria-controls={`dropdown-content-${id}`}
        className={cn(
          'dropdown-trigger flex items-center justify-between gap-2',
          {
            '!outline-0': position === 'overlay' && alignment === 'overlay',
          },
          buttonClassName
        )}
        id={`dropdown-button-${id}`}
        type="button"
        variant={variant}
        onClick={() => {
          if (!isOpen) {
            setIsOpen(true)
            focusFirstItem()
          }
        }}
        onFocus={() => {
          if (!isOpen) setIsOpen(true)
        }}
        onKeyDown={(e) => {
          if (e.key === 'ArrowDown') {
            e.preventDefault()
            if (!isOpen) setIsOpen(true)
            focusFirstItem()
          }
          if (e.key === 'Escape' && isOpen) {
            e.preventDefault()
            handleClose()
          }
        }}
        {...props}
      >
        {!hideIcon && iconAlignment === 'left' ? icon : null}
        {buttonText}
        {!hideIcon && iconAlignment === 'right' ? icon : null}
      </Button>

      {createPortal(dropdownContent, document.body)}
    </div>
  )
}
