'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { CaretUpDown } from '@phosphor-icons/react'
import { Button, IButton } from '@/components'

export interface IDropdown extends IButton {
  alignment?: 'left' | 'right' | 'overlay'
  children: React.ReactNode
  id: string
  isFullWidth?: boolean
  position?: 'above' | 'below' | 'beside' | 'overlay'
  text: React.ReactNode
}

export const Dropdown: FC<IDropdown> = ({
  alignment = 'left',
  className,
  children,
  hasCustomPadding = false,
  id,
  isFullWidth = false,
  position = 'below',
  text,
}) => {
  return (
    <>
      <div
        className={classNames('z-10 relative inline-block text-left group', {
          'w-full': isFullWidth,
        })}
        id={id}
        tabIndex={0}
      >
        <Button
          aria-haspopup="true"
          aria-expanded="true"
          aria-controls={`dropdown-content-${id}`}
          className={classNames('h-full bg-white dark:bg-black', {
            'px-4 py-2': hasCustomPadding,
            [`${className}`]: Boolean(className),
          })}
          hasCustomPadding={hasCustomPadding}
          id={`dropdown-button-${id}`}
          type="button"
        >
          <div className="flex items-center justify-between">
            {text}

            <CaretUpDown />
          </div>
        </Button>

        <div className="hidden group-focus-within:block w-inherit">
          <div
            className={classNames(
              'absolute border divide-y rounded-md shadow-md outline-none bg-white text-gray-950 dark:bg-black dark:text-gray-50',
              {
                'left-0': alignment === 'left' && position !== 'beside',
                'right-0': alignment === 'right' && position !== 'beside',
                'bottom-full mb-2': position === 'above',
                'mt-2': position === 'below',
                'top-0': position === 'beside',
                'right-full mr-2':
                  position === 'beside' && alignment === 'left',
                'left-full ml-2':
                  position === 'beside' && alignment === 'right',
                'top-0 w-inherit':
                  position === 'overlay' && alignment === 'overlay',
              }
            )}
            aria-labelledby={`dropdown-button-${id}`}
            id={`dropdown-content-${id}`}
          >
            {children}
          </div>
        </div>
      </div>
    </>
  )
}
