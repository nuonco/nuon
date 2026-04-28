import React, { type FC } from 'react'
import { cn } from '@/utils/classnames'

interface IPageSection extends React.HTMLAttributes<HTMLDivElement> {
  flush?: boolean
  isScrollable?: boolean
}

export const PageSection: FC<IPageSection> = ({
  className,
  children,
  flush = false,
  isScrollable: _isScrollable,
  ...props
}) => {
  return (
    <section
      className={cn(
        'w-full min-w-0 flex flex-col',
        flush ? '' : 'p-4 md:p-6 gap-4 md:gap-6',
        className
      )}
      {...props}
    >
      {children}
    </section>
  )
}
