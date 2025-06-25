import React, { forwardRef } from 'react'
import { cn } from '@/stratus/components/helpers'
import './Button.css'

export type TButtonSize = 'lg' | 'md' | 'sm' | 'xs'
export type TButtonVariant = 'danger' | 'ghost' | 'primary' | 'secondary'

export interface IButton extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  size?: TButtonSize
  variant?: TButtonVariant
}

export const Button = forwardRef<HTMLButtonElement, IButton>(
  (
    { className, children, size = 'md', variant = 'secondary', ...props },
    ref
  ) => {
    return (
      <button className={cn(variant, size, className)} ref={ref} {...props}>
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'
