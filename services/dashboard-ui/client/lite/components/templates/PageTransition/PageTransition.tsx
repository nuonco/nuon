import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/utils/classnames'

export interface IPageTransition
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  children: ReactNode
}

export const PageTransition = ({
  children,
  className,
  ...props
}: IPageTransition) => (
  <div
    data-page-transition
    className={cn('flex w-full min-w-0 flex-col', className)}
    {...props}
  >
    {children}
  </div>
)
