import classNames from 'classnames'
import React, { type FC } from 'react'
import { GoInfo } from 'react-icons/go'

export interface IToolTip {
  children: React.ReactNode
  position?: 'bottom' | 'top'
  tipContent: React.ReactNode
}

export const ToolTip: FC<IToolTip> = ({
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
        `}
      </style>
      <span className={classNames('tooltip')}>
        <span
          className={classNames('tooltip-wrapper', {
            [`${position}`]: true,
          })}
        >
          <div
            className={classNames(
              'bg-dark text-light dark:bg-light dark:text-dark text-xs p-2 rounded drop-shadow-md'
            )}
          >
            {tipContent}
          </div>
        </span>
        <span className="flex items-center gap-2 text-xs">
          {children}
          <GoInfo className="text-gray-600 dark:text-gray-400" />
        </span>
      </span>
    </>
  )
}
