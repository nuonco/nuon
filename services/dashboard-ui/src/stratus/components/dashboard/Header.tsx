import classNames from 'classnames'
import React, { type FC } from 'react'

interface IHeader extends React.HTMLAttributes<HTMLDivElement> {}

export const Header: FC<IHeader> = ({
  className,
  children,
  ...props
}) => {
  return (
    <header
      className={classNames('flex shrink-0 items-start justify-between p-4 md:p-6 md:h-28 w-full', {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </header>
  )
}
