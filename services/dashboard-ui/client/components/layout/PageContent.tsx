import React from 'react'
import { cn } from '@/utils/classnames'

interface IPageContent extends React.HTMLAttributes<HTMLDivElement> {
  isScrollable?: boolean
  variant?: 'primary' | 'secondary' | 'tertiary'
}

export const PageContent = ({
  className,
  children,
  isScrollable = false,
  variant = 'primary',
  ...props
}: IPageContent) => {
  return (
    <div
      className={cn(
        'flex-1 flex flex-col',
        {
          'md:flex-row': variant !== 'primary',
          'h-full overflow-y-auto overflow-x-hidden': isScrollable,
        },
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
