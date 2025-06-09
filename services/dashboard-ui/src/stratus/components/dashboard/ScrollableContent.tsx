import classNames from 'classnames'
import React, { type FC } from 'react'

interface IScrollableContent extends React.HTMLAttributes<HTMLDivElement> {}

export const ScrollableContent: FC<IScrollableContent> = ({
  className,
  children,
  ...props
}) => {
  return (
    <div
      className={classNames('overflow-scroll w-full', {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </div>
  )
}
