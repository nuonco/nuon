import classNames from 'classnames'
import React, { type FC } from 'react'
import { Button, IButton } from '@/components'

export interface IDropdown extends IButton {
  alignment?: 'left' | 'right'
  children: React.ReactNode
  id: string
  position?: 'above' | 'below' | 'beside'
  text: React.ReactNode
}

export const Dropdown: FC<IDropdown> = ({
  alignment = 'left',
  className,
  children,
  id,
  position = 'below',
  text,
}) => {
  return (
    <>
      <div
        className="z-10 relative inline-block text-left h-[56px] group"
        id={id}
        tabIndex={0}
      >
        <Button
          aria-haspopup="true"
          aria-expanded="true"
          aria-controls={`dropdown-content-${id}`}
          className={classNames('h-full', {
            [`${className}`]: Boolean(className),
          })}
          id={`dropdown-button-${id}`}
          type="button"
        >
          <div className="flex items-center justify-between">            
            {text}
            <svg
              className="w-5 h-5 ml-2 -mr-1"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                clipRule="evenodd"
              ></path>
            </svg>
          </div>
        </Button>

        <div className="hidden group-focus-within:block">
          <div
            className={classNames(
              'absolute min-w-56 border divide-y rounded shadow-md outline-none bg-gray-50 text-gray-950 dark:bg-gray-950 dark:text-gray-50',
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
