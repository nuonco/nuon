import classNames from 'classnames'
import React, { type FC } from 'react'

const defaultStyles =
  'bg-white text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50 hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10'

const cautionStyles =
  'bg-fuchsia-700 hover:bg-fuchsia-600 focus:bg-fuchsia-600 active:bg-fuchsia-800' +
  'bg-fuchsia-700 hover:bg-fuchsia-600 focus:bg-fuchsia-600 active:bg-fuchsia-800'

export type TButtonVariant =
  | 'default'
  | 'primary'
  | 'ghost'
  | 'caution'
  | 'danger'

export interface IButton extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  hasCustomPadding?: boolean
  type?: 'button' | 'reset' | 'submit'
  variant?: TButtonVariant
}

export const Button: FC<IButton> = ({
  children,
  className,
  hasCustomPadding = false,
  variant = 'default',
  ...props
}) => {
  return (
    <button
      className={classNames(
        'rounded-md border focus:outline outline-1 outline-primary-500 dark:outline-primary-400',
        {
          [`${defaultStyles} border`]: variant === 'default',
          'bg-primary-600 hover:bg-primary-700 focus:bg-primary-700 active:bg-primary-900':
            variant === 'primary',
          [`${defaultStyles} border-transparent`]: variant === 'ghost',
          [`${defaultStyles} border-red-800 text-red-800 dark:border-red-500 dark:text-red-500`]:
            variant === 'caution',
          'bg-red-700 hover:bg-red-600': variant === 'danger',
          'text-gray-50 px-5 border-transparent font-medium':
            variant === 'primary' || variant === 'danger',
          'px-3 py-1.5': !hasCustomPadding,
          'cursor-not-allowed text-cool-grey-500 dark:text-cool-grey-600 hover:!bg-transparent':
            props.disabled && variant !== 'primary',
          'cursor-not-allowed !text-cool-grey-500 !bg-primary-900 hover:!bg-primary-900':
            props.disabled && variant === 'primary',
          'cursor-not-allowed !text-cool-grey-500 !bg-red-900 hover:!bg-red-900':
            props.disabled && variant === 'danger',
          [`${className}`]: Boolean(className),
        }
      )}
      {...props}
    >
      {children}
    </button>
  )
}
