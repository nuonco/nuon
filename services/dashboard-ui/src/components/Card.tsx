import classNames from 'classnames'
import React, { type FC } from 'react'

export interface ICard extends React.HTMLAttributes<HTMLDivElement> {}

export const Card: FC<ICard> = ({ className, children, ...props }) => (
  <div
    className={classNames(
      'p-4 rounded-lg bg-slate-100 dark:bg-slate-900 drop-shadow-sm flex flex-col gap-2 overflow-auto',
      {
        [`${className}`]: Boolean(className),
      }
    )}
    {...props}
  >
    {children}
  </div>
)
