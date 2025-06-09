import classNames from 'classnames'
import React, { type FC } from 'react'

interface IPageLayout extends React.HTMLAttributes<HTMLDivElement> {}

export const PageLayout: FC<IPageLayout> = ({
  className,
  children,
  ...props
}) => {
  return (
    <div
      className={classNames(
        'flex-auto flex flex-col md:flex-row max-w-full overflow-hidden',
        {
          [`${className}`]: Boolean(className),
        }
      )}
      {...props}
    >
      {children}
    </div>
  )
}
