import React from 'react'
import { cn } from '@/utils/classnames'

interface IPageHeader extends React.HTMLAttributes<HTMLDivElement> {
  flush?: boolean
}

export const PageHeader = ({
  className,
  children,
  flush = false,
  ...props
}: IPageHeader) => {
  return (
    <header
      className={cn(
        'flex flex-wrap gap-3 shrink-0 items-start justify-between w-full',
        flush ? '' : 'p-4 md:p-6 md:min-h-28',
        className
      )}
      {...props}
    >
      {children}
    </header>
  )
}
