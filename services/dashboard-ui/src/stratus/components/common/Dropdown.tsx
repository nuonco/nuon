'use client'

import classNames from 'classnames'
import React, { type FC, forwardRef, useEffect, useRef, useState } from 'react'
import { CaretDown } from '@phosphor-icons/react'
import { Button, IButton } from './Button'
import './Dropdown.css'

const useFocusOutside = (handler: () => void) => {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleFocusIn = (event: FocusEvent) => {
      const relatedTarget = event.relatedTarget as HTMLElement | null

      if (ref.current && !ref.current.contains(relatedTarget)) {
        handler() // Call the handler if focus moves outside
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

export interface IDropdown extends IButton {
  alignment?: 'left' | 'right' | 'overlay'
  buttonClassName?: string
  buttonText: React.ReactNode
  children: React.ReactNode
  dropdownClassName?: string
  hideIcon?: boolean
  icon?: React.ReactNode
  iconAlignment?: 'left' | 'right'
  id: string
  position?: 'above' | 'below' | 'beside' | 'overlay'
  wrapperClassName?: string
}

export const Dropdown: FC<IDropdown> = ({
  alignment = 'left',
  buttonText,
  buttonClassName,
  className,
  children,
  dropdownClassName,
  hideIcon = false,
  icon = <CaretDown />,
  iconAlignment = 'right',
  id,
  position = 'below',
  variant,
}) => {
  const [isOpen, setIsOpen] = useState(false)

  const handleClose = () => {
    setIsOpen(false)
  }

  const dropdownRef = useFocusOutside(handleClose)
  const contentRef = useClickOutside(handleClose)

  return (
    <>
      <div
        className={classNames('dropdown', {
          [`${className}`]: Boolean(className),
        })}
        id={id}
        ref={dropdownRef}
      >
        <Button
          aria-haspopup="true"
          aria-expanded="true"
          aria-controls={`dropdown-content-${id}`}
          className={classNames('dropdown-trigger', {
            '!outline-0': position === 'overlay' && alignment === 'overlay',
            [`${buttonClassName}`]: Boolean(buttonClassName),
          })}
          id={`dropdown-button-${id}`}
          type="button"
          variant={variant}
          onClick={() => {
            if (!isOpen) setIsOpen(true)
          }}
          onFocus={() => {
            if (!isOpen) setIsOpen(true)
          }}
        >
          {!hideIcon && iconAlignment === 'left' ? icon : null}
          {buttonText}
          {!hideIcon && iconAlignment === 'right' ? icon : null}
        </Button>

        <TransitionDiv
          ref={contentRef}
          className={classNames(`dropdown-content ${alignment} ${position}`, {
            [`${dropdownClassName}`]: Boolean(dropdownClassName),
          })}
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

interface IDropdownContent extends React.HTMLAttributes<HTMLDivElement> {
  isVisible: boolean
  onExited?: () => void
}

const TransitionDiv = forwardRef<HTMLDivElement, IDropdownContent>(
  ({ children, className, isVisible, onExited, ...props }, ref) => {
    const [isExiting, setIsExiting] = useState(false)
    const [isMounted, setIsMounted] = useState(isVisible)

    useEffect(() => {
      if (isVisible) {
        setIsMounted(true) // Mount the component
        setIsExiting(false) // Remove the exit class
      } else {
        setIsExiting(true) // Add the exit class
        const timeout = setTimeout(() => {
          setIsMounted(false) // Unmount the component after the animation
          onExited?.() // Notify parent that the component has exited
        }, 155) // Duration should match CSS animation duration

        return () => clearTimeout(timeout) // Cleanup timeout on unmount
      }
    }, [isVisible, onExited])

    if (!isMounted) {
      return null // Don't render anything if the component is not mounted
    }

    return (
      <div
        className={classNames(`${isExiting ? 'exit' : 'enter'}`, {
          [`${className}`]: Boolean(className),
        })}
        ref={ref}
        {...props}
      >
        {children}
      </div>
    )
  }
)

TransitionDiv.displayName = 'TransitionDiv'
