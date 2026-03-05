import React from 'react'
import { cn } from '@/utils/classnames'

interface IPageHeader extends React.HTMLAttributes<HTMLDivElement> {}

export const PageHeader = ({ className, children, ...props }: IPageHeader) => {
  return (
    <header
      className={cn(
        'flex flex-wrap gap-3 shrink-0 items-start justify-between p-4 md:p-6 md:min-h-28 w-full',
        className
      )}
      {...props}
    >
      {children}
    </header>
  )
}
