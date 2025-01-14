'use client'

import classNames from 'classnames'
import React, { type FC, useCallback, useEffect } from 'react'
import { X } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { Heading } from '@/components/Typography'

export interface IModal extends React.HTMLAttributes<HTMLDivElement> {
  actions?: React.ReactNode | null
  heading: React.ReactNode
  hasFixedHeight?: boolean
  isOpen?: boolean
  onClose?: () => void
}

export const Modal: FC<IModal> = ({
  className,
  children,
  actions = null,
  heading,
  hasFixedHeight = false,
  isOpen = false,
  onClose = () => {},
  ...props
}) => {
  const onEscape = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose()
    }
  }, [])

  useEffect(() => {
    document.addEventListener('keydown', onEscape, false)
    return () => {
      document.removeEventListener('keydown', onEscape, false)
    }
  }, [])

  return isOpen ? (
    <div className="absolute flex w-full h-full top-0 left-0 z-50">
      <div
        className="fixed bg-black/50 dark:bg-black/75 w-full h-full"
        onClick={onClose}
      />
      <div
        className={classNames(
          'relative z-50 border rounded-lg shadow-lg m-auto w-full max-w-7xl bg-white text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50',
          {
            [`${className}`]: Boolean(className),
          }
        )}
        {...props}
      >
        <header className="flex items-center justify-between px-6 py-4 border-b">
          <Heading>{heading}</Heading>
          <div className="flex items-center divide-x">
            {actions ? <div className="pr-4">{actions}</div> : null}
            <div className="pl-4">
              <Button className="!p-2" onClick={onClose}>
                <X />
              </Button>
            </div>
          </div>
        </header>
        <div
          className={classNames(
            'p-6 h-full max-h-[700px] overflow-y-auto overflow-x-hidden',
            {
              'min-h-[700px]': hasFixedHeight,
            }
          )}
        >
          {children}
        </div>
      </div>
    </div>
  ) : null
}
