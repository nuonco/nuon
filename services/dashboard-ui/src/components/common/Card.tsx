import classNames from 'classnames'
import React, { type FC } from 'react'

export interface ICard extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactElement
}

export const Card: FC<ICard> = ({ className, children, ...props }) => (
  <div
    className={classNames(
      'p-4 rounded-lg bg-gray-100 dark:bg-gray-900 drop-shadow-sm flex flex-col gap-4 overflow-auto',
      {
        [className]: Boolean(className),
      }
    )}
    {...props}
  >
    {children}
  </div>
)
