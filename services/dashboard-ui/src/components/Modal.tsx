'use client'

import classNames from 'classnames'
import React, { type FC, useCallback, useEffect, useRef, useState } from 'react'
import { X } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Heading } from '@/components/Typography'

export interface IModal extends React.HTMLAttributes<HTMLDivElement> {
  actions?: React.ReactNode | null
  heading: React.ReactNode
  hasFixedHeight?: boolean
  isOpen?: boolean
  onClose?: () => void
  contentClassName?: string
  showCloseButton?: boolean
}

export const Modal: FC<IModal> = ({
  className,
  children,
  actions = null,
  heading,
  hasFixedHeight = false,
  isOpen = false,
  onClose = () => {},
  contentClassName,
  showCloseButton = true,
  ...props
}) => {
  const modalRef = useRef(null)
  const [isVisible, setIsVisible] = useState(false)
  const [shouldRender, setShouldRender] = useState(false)

  const onEscape = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose()
    }
  }, [onClose])

  useEffect(() => {
    document.addEventListener('keydown', onEscape, false)
    return () => {
      document.removeEventListener('keydown', onEscape, false)
    }
  }, [onEscape])

  // Handle entrance and exit animations
  useEffect(() => {
    if (isOpen) {
      setShouldRender(true)
      // Trigger entrance animation after mounting
      const timer = setTimeout(() => setIsVisible(true), 10)
      return () => clearTimeout(timer)
    } else {
      setIsVisible(false)
      // Wait for exit animation before unmounting
      const timer = setTimeout(() => setShouldRender(false), 300)
      return () => clearTimeout(timer)
    }
  }, [isOpen])

  useEffect(() => {
    if (isVisible) {
      modalRef?.current?.focus()
    }
  }, [isVisible])

  return shouldRender ? (
    <div className="absolute flex w-full h-full top-0 left-0 z-[60]">
      {/* Backdrop with fade transition */}
      <div
        className={classNames(
          'fixed w-full h-full transition-opacity duration-300 ease-out',
          isVisible
            ? 'bg-black/50 dark:bg-black/75 opacity-100'
            : 'bg-black/50 dark:bg-black/75 opacity-0'
        )}
        onClick={onClose}
      />
      {/* Modal with scale + fade transition */}
      <div
        className={classNames(
          'relative z-[60] border rounded-lg shadow-lg m-auto w-full bg-white text-cool-grey-950 dark:bg-dark-grey-900 dark:text-cool-grey-50 focus:outline outline-1 outline-primary-500 dark:outline-white/40 transition-all duration-300 ease-out transform max-w-[calc(100vw_-_4rem)]',
          isVisible
            ? 'opacity-100 scale-100 translate-y-0'
            : 'opacity-0 scale-95 translate-y-4',
          {
            [`${className}`]: Boolean(className),
          }
        )}
        {...props}
        tabIndex={-1}
        ref={modalRef}
      >
        <header className="flex items-center justify-between px-6 py-4 border-b">
          <Heading>{heading}</Heading>
          <div className="flex items-center divide-x">
            {actions ? (
              <div className={showCloseButton ? 'pr-4' : ''}>{actions}</div>
            ) : null}
            {showCloseButton && (
              <div className="pl-4">
                <Button className="!p-2" onClick={onClose}>
                  <X />
                </Button>
              </div>
            )}
          </div>
        </header>
        <div
          tabIndex={-1}
          className={classNames(
            'p-6 h-full max-h-[700px] overflow-y-auto overflow-x-hidden rounded-b-lg focus:outline outline-1 outline-primary-500 dark:outline-white/20',
            {
              'min-h-[700px]': hasFixedHeight,
              [`${contentClassName}`]: Boolean(contentClassName),
            }
          )}
        >
          {children}
        </div>
      </div>
    </div>
  ) : null
}
