import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'

export interface IShellBackground
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {}

export const ShellBackground = ({ className, ...props }: IShellBackground) => (
  <div
    data-shell-background
    aria-hidden="true"
    className={cn(
      'shell-background pointer-events-none absolute inset-0 z-0',
      className
    )}
    {...props}
  />
)
