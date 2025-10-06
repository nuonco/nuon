'use client'

import React, { useEffect, useRef, useState } from 'react'
import { cn } from '@/utils/classnames'
import { Button, type IButtonAsButton } from './Button'
import { Icon } from './Icon'
import { TransitionDiv } from './TransitionDiv'
import './Dropdown.css'

const useFocusOutside = (handler: () => void) => {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleFocusIn = (event: FocusEvent) => {
      const relatedTarget = event.relatedTarget as HTMLElement | null

      if (ref.current && !ref.current.contains(relatedTarget)) {
        handler()
      }
    }

    ref?.current?.addEventListener('focusout', handleFocusIn, true)

    return () => {
      ref?.current?.removeEventListener('focusout', handleFocusIn, true)
    }
  }, [handler])

  return ref
}

const useClickOutside = (handler: () => void) => {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        handler()
      }
    }

    document.addEventListener('mousedown', handleClickOutside)

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [handler])

  return ref
}

const getDropdownContentPositionClasses = (
  position: NonNullable<IDropdown['position']>,
  alignment: NonNullable<IDropdown['alignment']>
) => {
  // Static positioning (absolute, etc.)
  const classes = [
    'absolute',
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
    'z-[1000]',
  ]
  // Alignment and position
  if (position === 'below') {
    classes.push('mt-2')
    if (alignment === 'left') classes.push('left-0')
    if (alignment === 'right') classes.push('right-0')
  }
  if (position === 'above') {
    classes.push('bottom-full', 'mb-2')
    if (alignment === 'left') classes.push('left-0')
    if (alignment === 'right') classes.push('right-0')
  }
  if (position === 'beside') {
    classes.push('top-0')
    if (alignment === 'left') classes.push('right-full', 'mr-2')
    if (alignment === 'right') classes.push('left-full', 'ml-2')
  }
  if (position === 'overlay') {
    classes.push('top-0')
    // overlay is always positioned on top
  }
  return classes.join(' ')
}

export interface IDropdown extends IButtonAsButton {
  alignment?: 'left' | 'right' | 'overlay'
  buttonClassName?: string
  buttonText: React.ReactNode
  children: React.ReactNode
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
  className,
  children,
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

  const handleClose = () => {
    setIsOpen(false)
  }

  const dropdownRef = useFocusOutside(handleClose)
  const contentRef = useClickOutside(handleClose)

  return (
    <>
      <div
        className={cn(
          'dropdown relative inline-block text-left leading-none',
          className
        )}
        id={id}
        ref={dropdownRef}
      >
        <Button
          aria-haspopup="true"
          aria-expanded="true"
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

        <TransitionDiv
          ref={contentRef}
          className={cn(
            'dropdown-content',
            getDropdownContentPositionClasses(position, alignment),
            alignment,
            position,
            dropdownClassName
          )}
          aria-labelledby={`dropdown-button-${id}`}
          id={`dropdown-content-${id}`}
          isVisible={isOpen}
          tabIndex={-1}
        >
          {children}
        </TransitionDiv>
      </div>
    </>
  )
}
