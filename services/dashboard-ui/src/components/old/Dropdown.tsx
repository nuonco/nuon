'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { CaretUpDownIcon, CaretDownIcon } from '@phosphor-icons/react'
import { Button, IButton } from '@/components/old/Button'

export interface IDropdown extends IButton {
  alignment?: 'left' | 'right' | 'overlay'
  children: React.ReactNode
  id: string
  isFullWidth?: boolean
  position?: 'above' | 'below' | 'beside' | 'overlay'
  text: React.ReactNode
  dropdownContentClassName?: string
  wrapperClassName?: string
  isDownIcon?: boolean
  noIcon?: boolean
}

export const Dropdown: FC<IDropdown> = ({
  alignment = 'left',
  className,
  children,
  disabled = false,
  hasCustomPadding = false,
  id,
  isFullWidth = false,
  position = 'below',
  text,
  variant,
  dropdownContentClassName,
  wrapperClassName,
  isDownIcon = false,
  noIcon = false,
}) => {
  return (
    <>
      <div
        className={classNames(
          'relative inline-block text-left group leading-none',
          {
            'w-full': isFullWidth,
            [`${wrapperClassName}`]: Boolean(wrapperClassName),
          }
        )}
        id={id}
        tabIndex={0}
      >
        <Button
          aria-haspopup="true"
          aria-expanded="true"
          aria-controls={`dropdown-content-${id}`}
          className={classNames('h-full', {
            'px-4 py-2': hasCustomPadding,
            'group-focus-within:opacity-0':
              position === 'overlay' && alignment === 'overlay',
            [`${className}`]: Boolean(className),
          })}
          disabled={disabled}
          hasCustomPadding={hasCustomPadding}
          id={`dropdown-button-${id}`}
          type="button"
          variant={variant}
        >
          <div className="flex items-center justify-between gap-2 w-full">
            {text}

            {noIcon ? null : variant !== 'ghost' && isDownIcon ? (
              <CaretDownIcon />
            ) : (
              <CaretUpDownIcon />
            )}
          </div>
        </Button>

        <div className="hidden group-focus-within:block w-inherit">
          <div
            className={classNames(
              'absolute z-20 border divide-y rounded-md shadow-md outline-none bg-white text-cool-grey-950 dark:bg-dark-grey-900 dark:text-cool-grey-50',
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
                [`${dropdownContentClassName}`]: Boolean(
                  dropdownContentClassName
                ),
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
