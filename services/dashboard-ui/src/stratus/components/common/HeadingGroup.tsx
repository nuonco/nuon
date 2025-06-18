import classNames from 'classnames'
import React, { type FC } from 'react'

interface IHeadingGroup extends React.HTMLAttributes<HTMLDivElement> {}

export const HeadingGroup: FC<IHeadingGroup> = ({
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
