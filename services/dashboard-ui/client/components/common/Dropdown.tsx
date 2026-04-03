import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
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
  icon = <Icon variant="CaretDown" />,
  iconAlignment = 'right',
  id,
  isOpen: initIsOpen = false,
  position = 'below',
  variant,
  ...props
}: IDropdown) => {
  const [isOpen, setIsOpen] = useState(initIsOpen)
  const [styles, setStyles] = useState<React.CSSProperties>({})
  const triggerRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement | null>(null)
  const childPortals = useRef<Set<HTMLElement>>(new Set())
  const parentNesting = useContext(DropdownNestingContext)

  const handleClose = useCallback(() => {
    setIsOpen(false)
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

      if (parentNesting) {
        if (prev) parentNesting.unregisterChild(prev)
        if (el) parentNesting.registerChild(el)
      }
    },
    [parentNesting]
  )

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
    const newStyles: React.CSSProperties = {
      position: 'fixed',
      zIndex: 60,
    }

    if (position === 'below') {
      newStyles.top = trigger.bottom + 8
      if (alignment === 'left') newStyles.left = trigger.left
      if (alignment === 'right') newStyles.right = window.innerWidth - trigger.right
    }

    if (position === 'above') {
      newStyles.bottom = window.innerHeight - trigger.top + 8
      if (alignment === 'left') newStyles.left = trigger.left
      if (alignment === 'right') newStyles.right = window.innerWidth - trigger.right
    }

    if (position === 'beside') {
      newStyles.top = trigger.top
      if (alignment === 'left') newStyles.right = window.innerWidth - trigger.left + 8
      if (alignment === 'right') newStyles.left = trigger.right + 8
    }

    if (position === 'overlay') {
      newStyles.top = trigger.top
      newStyles.left = trigger.left
    }

    setStyles(newStyles)
  }, [position, alignment])

  useEffect(() => {
    if (!isOpen) return

    calculatePosition()

    window.addEventListener('resize', calculatePosition)
    window.addEventListener('scroll', calculatePosition, true)
    return () => {
      window.removeEventListener('resize', calculatePosition)
      window.removeEventListener('scroll', calculatePosition, true)
    }
  }, [isOpen, calculatePosition])

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
    const handleFocusOut = (event: FocusEvent) => {
      if (!isInsideTree(event.relatedTarget as Node)) {
        handleClose()
      }
    }

    triggerEl?.addEventListener('focusout', handleFocusOut, true)
    return () => {
      triggerEl?.removeEventListener('focusout', handleFocusOut, true)
    }
  }, [isOpen, closeOnBlur, handleClose, isInsideTree])

  const dropdownContent = (
    <TransitionDiv
      ref={contentCallbackRef}
      className={cn(
        'dropdown-content',
        'border',
        'divide-y',
        'rounded-md',
        'shadow-md',
        'outline-none',
        'bg-white',
        'dark:bg-dark-grey-900',
        'w-fit',
        'outline-offset-0',
        'outline-1',
        'outline-primary-400/10',
        alignment,
        position,
        dropdownClassName
      )}
      aria-labelledby={`dropdown-button-${id}`}
      id={`dropdown-content-${id}`}
      isVisible={isOpen}
      style={styles}
      tabIndex={-1}
      onClick={(e) => {
        if (!closeOnBlur) return
        const target = e.target as HTMLElement
        if (target.closest('button, a, [role="menuitem"]')) {
          handleClose()
        }
      }}
    >
      <DropdownNestingContext.Provider value={nestingContext}>
        {children}
      </DropdownNestingContext.Provider>
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
          'dropdown-trigger flex items-center justify-between gap-2 h-fit focus:outline-primary-400/80',
          {
            '!outline-0': position === 'overlay' && alignment === 'overlay',
          },
          buttonClassName
        )}
        id={`dropdown-button-${id}`}
        type="button"
        variant={variant}
        onClick={() => {
          if (!isOpen) setIsOpen(true)
        }}
        onFocus={() => {
          if (!isOpen) setIsOpen(true)
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
