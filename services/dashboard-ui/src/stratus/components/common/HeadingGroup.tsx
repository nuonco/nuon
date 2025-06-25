import React from 'react'
import { cn } from '@/stratus/components/helpers'

interface IHeadingGroup extends React.HTMLAttributes<HTMLDivElement> {}

export const HeadingGroup = ({
  className,
  children,
  ...props
}: IHeadingGroup) => {
  return (
    <hgroup className={cn('flex flex-col', className)} {...props}>
      {children}
    </hgroup>
  )
}
