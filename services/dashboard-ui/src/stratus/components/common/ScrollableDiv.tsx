import React from 'react'
import { cn } from '@/stratus/components/helpers'

interface IScrollableDiv extends React.HTMLAttributes<HTMLDivElement> {}

export const ScrollableDiv = ({
  className,
  children,
  ...props
}: IScrollableDiv) => {
  return (
    <div
      className={cn('overflow-y-auto w-full max-w-full', className)}
      {...props}
    >
      {children}
    </div>
  )
}
