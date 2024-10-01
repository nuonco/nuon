'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { Info } from '@phosphor-icons/react'

export interface IToolTip {
  alignment?: 'center' | 'left' | 'right'
  children: React.ReactNode
  position?: 'bottom' | 'top'
  tipContent: React.ReactNode
}

export const ToolTip: FC<IToolTip> = ({
  alignment = 'left',
  children,
  position = 'top',
  tipContent,
}) => {
  return (
    <>
      <style>
        {`
          .tooltip {
            display: inline-block;
            position: relative;
            width: fit-content;
            z-index: 10;
          }

          .tooltip .tooltip-wrapper {
            display: none;
            height: max-content;
            position: absolute;
            width: max-content;
            z-index: 10;
          }

          .tooltip:hover .tooltip-wrapper {
            display: block;
          }

          .tooltip .tooltip-wrapper.top {
            padding-bottom: 0.5rem;
          }

          .tooltip:hover .tooltip-wrapper.top {
            top: 0;
            transform: translateY(-100%);
          }

          .tooltip .tooltip-wrapper.bottom {
            padding-top: 0.5rem;
          }

          .tooltip:hover .tooltip-wrapper.bottom {
            bottom: 0;
            transform: translateY(100%);
          }

          .tooltip:hover .tooltip-wrapper.right {
            right: 0;
          }

          .tooltip:hover .tooltip-wrapper.left {
            left: 0;
          }

          .tooltip:hover .tooltip-wrapper.center {
            left: 50%;
            transform: translateX(-50$);
          }
        `}
      </style>
      <span className={classNames('tooltip')}>
        <span
          className={classNames('tooltip-wrapper', {
            [`${alignment}`]: true,
            [`${position}`]: true,
          })}
        >
          <div
            className={classNames(
              'bg-dark text-light dark:bg-light dark:text-dark text-sm px-2 py-1.5 rounded drop-shadow-md max-w-96 '
            )}
          >
            {tipContent}
          </div>
        </span>
        <span className="flex items-center gap-2 text-sm">
          {children}
          <Info className="text-cool-grey-600 dark:text-cool-grey-500" />
        </span>
      </span>
    </>
  )
}
