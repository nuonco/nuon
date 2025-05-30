'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { CaretDown } from '@phosphor-icons/react'
import { Button, IButton } from './Button'

export interface IDropdown extends IButton {
  alignment?: 'left' | 'right' | 'overlay'
  buttonClassName?: string
  buttonText: React.ReactNode 
  children: React.ReactNode
  dropdownClassName?: string
  icon?: React.ReactNode
  id: string
  position?: 'above' | 'below' | 'beside' | 'overlay'
  wrapperClassName?: string
  isDownIcon?: boolean
}

export const Dropdown: FC<IDropdown> = ({
  alignment = 'left',
  buttonText,
  buttonClassName,
  className,
  children,
  dropdownClassName,
  icon = <CaretDown />,
  id,
  position = 'below',
  variant,
}) => {
  return (
    <>
      <div
        className={classNames(
          'z-10 relative inline-block text-left group leading-none dropdown',
          {
            [`${className}`]: Boolean(className),
          }
        )}
        id={id}
        tabIndex={0}
      >
        <Button
          aria-haspopup="true"
          aria-expanded="true"
          aria-controls={`dropdown-content-${id}`}
          className={classNames('!h-fit', {
            'group-focus-within:opacity-0':
              position === 'overlay' && alignment === 'overlay',
            [`${buttonClassName}`]: Boolean(buttonClassName),
          })}
          id={`dropdown-button-${id}`}
          type="button"
          variant={variant}
        >
          <div className="flex items-center justify-between gap-2">
            {buttonText}
            {icon}
          </div>
        </Button>

        <div className="hidden group-focus-within:block w-inherit">
          <div
            className={classNames(
              'absolute z-20 border divide-y rounded-md shadow-md outline-none bg-white dark:bg-dark-grey-100 w-fit',
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
                'top-0 w-fit':
                  position === 'overlay' && alignment === 'overlay',
                [`${dropdownClassName}`]: Boolean(
                  dropdownClassName
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
