import classNames from 'classnames'
import React, { forwardRef } from 'react'
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
      <button
        className={classNames(`${variant} ${size}`, {
          [`${className}`]: Boolean(className),
        })}
        ref={ref}
        {...props}
      >
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'
