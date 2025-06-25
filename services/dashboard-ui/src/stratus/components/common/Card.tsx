import React from 'react'
import { cn } from '@/stratus/components/helpers'

export interface ICard extends React.HTMLAttributes<HTMLDivElement> {}

export const Card = ({ children, className, ...props }: ICard) => {
  return (
    <div
      className={cn('flex flex-col gap-6 p-6 border rounded-md', className)}
      {...props}
    >
      {children}
    </div>
  )
}
