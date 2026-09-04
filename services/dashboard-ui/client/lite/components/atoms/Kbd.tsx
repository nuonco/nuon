import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/utils/classnames'

export interface IKbd extends HTMLAttributes<HTMLElement> {
  children: ReactNode
}

export const Kbd = ({ className, children, ...props }: IKbd) => (
  <kbd
    className={cn(
      'inline-flex h-4 min-w-4 items-center justify-center gap-0.5 rounded bg-surface-03 px-1 font-sans text-[0.625rem] leading-none text-secondary',
      className
    )}
    {...props}
  >
    {children}
  </kbd>
)
