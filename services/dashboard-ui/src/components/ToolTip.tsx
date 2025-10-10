'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { Info } from '@phosphor-icons/react'

export interface IToolTip {
  parentClassName?: string
  alignment?: 'center' | 'left' | 'right'
  children: React.ReactNode
  isIconHidden?: boolean
  position?: 'bottom' | 'top'
  tipContent: React.ReactNode
}

export const ToolTip: FC<IToolTip> = ({
  alignment = 'left',
  children,
  isIconHidden = false,
  parentClassName,
  position = 'top',
  tipContent,
}) => {
  return (
    <>
      <span className={classNames('tooltip', {
        [`${parentClassName}`]: Boolean(parentClassName),
      })}>
        <span
          className={classNames('tooltip-wrapper !z-20', {
            [`${alignment}`]: true,
            [`${position}`]: true,
          })}
        >
          <div
            className={classNames(
              'bg-dark-grey-900 text-white dark:bg-white dark:text-dark-grey-900 text-sm px-2 py-1.5 rounded drop-shadow-md max-w-96'
            )}
          >
            {tipContent}
          </div>
        </span>
        <span className="flex items-center gap-1 text-xs">
          {children}
          {isIconHidden ? null : (
            <Info className="text-cool-grey-600 dark:text-cool-grey-500" />
          )}
        </span>
      </span>
    </>
  )
}
