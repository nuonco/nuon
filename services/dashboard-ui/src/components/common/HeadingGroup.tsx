import React from 'react'
import { cn } from '@/utils/classnames'

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
