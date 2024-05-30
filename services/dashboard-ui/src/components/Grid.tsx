import React, { type FC } from 'react'
import classNames from 'classnames'

export interface IGrid extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode
  variant?: 'default' | '3-cols'
}

export const Grid: FC<IGrid> = ({ className, children, variant = 'default' }) => {
  return (
    <div
      className={classNames('grid gap-6 w-full', {
        'auto-rows-auto grid-cols-auto':
          variant === 'default',
        'grid-cols-1 md:grid-cols-2 lg:grid-cols-3  h-fit overflow-hidden':
        variant === '3-cols',
         [`${className}`]: Boolean(className),
      })}
    >
      {children}
    </div>
  )
}
