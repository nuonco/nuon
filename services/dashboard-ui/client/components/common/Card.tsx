import React from 'react'
import { cn } from '@/utils/classnames'

export interface ICard extends React.HTMLAttributes<HTMLDivElement> {}

export const Card = ({ children, className, ...props }: ICard) => {
  return (
    <div
      className={cn(
        'flex flex-col gap-6 p-6 border rounded-md shadow-sm',
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}
