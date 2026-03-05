import React, { type FC } from 'react'
import { cn } from '@/utils/classnames'

interface IPageGrid extends React.HTMLAttributes<HTMLDivElement> {}

export const PageGrid: FC<IPageGrid> = ({ className, children, ...props }) => {
  return (
    <div
      className={cn(
        'w-full grid grid-cols-1 md:grid-cols-[1fr_372px]',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
