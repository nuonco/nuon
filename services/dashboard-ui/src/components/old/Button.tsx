import classNames from 'classnames'
import React, { forwardRef } from 'react'

const defaultStyles =
  'bg-white text-cool-grey-950 dark:bg-dark-grey-900 dark:text-cool-grey-50 hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10 text-[13px] hover:cursor-pointer'

const cautionStyles =
  'bg-fuchsia-700 hover:bg-fuchsia-600 focus:bg-fuchsia-600 active:bg-fuchsia-800' +
  'bg-fuchsia-700 hover:bg-fuchsia-600 focus:bg-fuchsia-600 active:bg-fuchsia-800'

export type TButtonVariant =
  | 'default'
  | 'primary'
  | 'ghost'
  | 'caution'
  | 'danger'
  | 'secondary'
| 'menu'

export interface IButton extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  hasCustomPadding?: boolean
  type?: 'button' | 'reset' | 'submit'
  variant?: TButtonVariant
  ref?: any
}

export const Button = forwardRef<HTMLButtonElement, IButton>(
  (
    {
      children,
      className,
      hasCustomPadding = false,
      variant = 'default',
      ...props
    },
    ref
  ) => {
    return (
      <button
        className={classNames(
          'rounded-md border font-sans focus:outline focus:outline-1 focus:outline-primary-500 dark:focus:outline-primary-400 text-nowrap !shadow-none !grow-0 h-auto',
          {
            [`${defaultStyles} border`]: variant === 'default',
            'bg-primary-600 hover:bg-primary-700 focus:bg-primary-700 active:bg-primary-900':
              variant === 'primary',
            [`${defaultStyles} !border-transparent`]: variant === 'ghost',
            [`${defaultStyles} border-red-800 text-red-800 dark:border-red-500 dark:text-red-500`]:
              variant === 'caution',
            [`${defaultStyles} border bg-cool-grey-300 text-primary-400 dark:bg-dark-grey-600 dark:text-primary-400`]: variant === 'secondary',
            'bg-red-700 hover:bg-red-600': variant === 'danger',
            'text-gray-50 px-5 border-transparent font-medium':
              variant === 'primary' || variant === 'danger',
            'px-3 py-1.5': !hasCustomPadding,
            'cursor-not-allowed !text-cool-grey-500 dark:!text-cool-grey-600 !bg-white dark:!bg-dark-grey-700':
              props.disabled && variant !== 'primary' && variant !== 'danger',
            'cursor-not-allowed !text-cool-grey-500 !bg-primary-900 hover:!bg-primary-900':
              props.disabled && variant === 'primary',
            'cursor-not-allowed !text-cool-grey-500 !bg-red-900 hover:!bg-red-900':
            props.disabled && variant === 'danger',
            [`${defaultStyles} border-transparent text-sm !font-medium !py-2 !px-3 h-[36px] flex items-center gap-3 w-full`]: variant === 'menu',
            "hover:!cursor-pointer": true,
            [`${className}`]: Boolean(className),
          }
        )}
        ref={ref}
        {...props}
      >
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'
