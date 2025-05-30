import classNames from 'classnames'
import React, { type FC } from 'react'

interface IPageContent extends React.HTMLAttributes<HTMLDivElement> {}

export const PageContent: FC<IPageContent> = ({
  className,
  children,
  ...props
}) => {
  return (
    <div
      className={classNames(
        'flex-auto flex flex-col md:flex-row max-w-full overflow-y-auto',
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
