import classNames from 'classnames'
import React, { type FC } from 'react'

interface IHeaderGroup extends React.HTMLAttributes<HTMLDivElement> {}

export const HeaderGroup: FC<IHeaderGroup> = ({
  className,
  children,
  ...props
}) => {
  return (
    <hgroup
      className={classNames('flex flex-col', {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </hgroup>
  )
}
