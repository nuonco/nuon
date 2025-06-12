'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useRef, useState } from 'react'
import { CaretDown } from '@phosphor-icons/react'
import { Button, IButton } from './Button'
import { TransitionDiv } from "./TransitionDiv"
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
