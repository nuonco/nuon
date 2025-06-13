import classNames from 'classnames'
import React, { type FC } from 'react'

interface IScrollableDiv extends React.HTMLAttributes<HTMLDivElement> {}

export const ScrollableDiv: FC<IScrollableDiv> = ({
  className,
  children,
  ...props
}) => {
  return (
    <div
      className={classNames('overflow-y-auto w-full max-w-full', {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </div>
  )
}
