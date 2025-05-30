import classNames from 'classnames'
import React, { type FC } from 'react'

interface IPageHeader extends React.HTMLAttributes<HTMLDivElement> {}

export const PageHeader: FC<IPageHeader> = ({
  className,
  children,
  ...props
}) => {
  return (
    <header
      className={classNames('px-4 py-4 md:px-8 md:py-6 flex flex-col border-b', {
        [`${classNames}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </header>
  )
}
