import React, { type FC } from 'react'
import { cn } from '@/utils/classnames'

interface IPageSection extends React.HTMLAttributes<HTMLDivElement> {
  isScrollable?: boolean
}

export const PageSection: FC<IPageSection> = ({
  className,
  children,
  isScrollable = false,
  ...props
}) => {
  return (
    <section
      className={cn(
        'p-4 md:p-6 w-full flex flex-col gap-4 md:gap-6',
        {
          'h-full overflow-y-auto overflow-x-hidden': isScrollable,
        },
        className
      )}
      {...props}
    >
      {children}
    </section>
  )
}
