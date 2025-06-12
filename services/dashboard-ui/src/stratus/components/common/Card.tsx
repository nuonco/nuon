import classNames from 'classnames'
import React, { type FC } from 'react'

export interface ICard extends React.HTMLAttributes<HTMLDivElement> {}

export const Card: FC<ICard> = ({ children, className, ...props }) => {
  return (
    <div
      className={classNames('flex flex-col gap-6 p-6 border rounded-md', {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </div>
  )
}
