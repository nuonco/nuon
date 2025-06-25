import React from 'react'
import { cn } from '@/stratus/components/helpers'

interface IPageLayout extends React.HTMLAttributes<HTMLDivElement> {}

export const PageLayout = ({ className, children, ...props }: IPageLayout) => {
  return (
    <div
      className={cn(
        'flex-auto flex flex-col md:flex-row max-w-full overflow-y-auto md:overflow-hidden',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
