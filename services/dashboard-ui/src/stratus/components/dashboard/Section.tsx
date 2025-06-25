import React, { type FC } from 'react'
import { cn } from '@/stratus/components/helpers'

interface ISection extends React.HTMLAttributes<HTMLDivElement> {}

export const Section: FC<ISection> = ({ className, children, ...props }) => {
  return (
    <section
      className={cn('p-4 md:p-6 w-full flex flex-col', className)}
      {...props}
    >
      {children}
    </section>
  )
}
