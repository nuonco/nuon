import classNames from 'classnames'
import React, { type FC } from 'react'
import './Button.css'

export type TButtonSize = 'lg' | 'md' | 'sm' | 'xs'
export type TButtonVariant = 'danger' | 'ghost' | 'primary' | 'secondary'

export interface IButton extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  size?: TButtonSize
  variant?: TButtonVariant
}

export const Button: FC<IButton> = ({
  className,
  children,
  size = 'md',
  variant = 'secondary',
  ...props
}) => {
  return (
    <button
      className={classNames(`${variant} ${size}`, {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </button>
  )
}
