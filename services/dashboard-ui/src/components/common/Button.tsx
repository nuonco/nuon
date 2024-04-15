import classNames from 'classnames'
import React, { type FC } from 'react'

const defaultStyles =
  'bg-gray-50 hover:bg-gray-100 focus:bg-gray-100 active:bg-gray-200 text-gray-950 dark:bg-gray-950 dark:hover:bg-gray-900 dark:focus:bg-gray-900 dark:active:bg-gray-800 dark:text-gray-50'

const cautionStyles =
  'bg-fuchsia-700 hover:bg-fuchsia-600 focus:bg-fuchsia-600 active:bg-fuchsia-800' +
  'bg-fuchsia-700 hover:bg-fuchsia-600 focus:bg-fuchsia-600 active:bg-fuchsia-800'

export type TButtonVariant =
  | 'default'
  | 'primary'
  | 'ghost'
  | 'caution'
  | 'danger'

export interface IButton extends React.HTMLAttributes<HTMLButtonElement> {
  variant?: TButtonVariant
}

export const Button: FC<IButton> = ({
  children,
  className,
  variant = 'default',
  ...props
}) => {
  return (
    <button
      className={classNames('px-3 py-1.5 rounded-sm border', {
        [`${defaultStyles} border-fuchsia-500`]: variant === 'default',
        'bg-fuchsia-700 hover:bg-fuchsia-600 focus:bg-fuchsia-600 active:bg-fuchsia-800':
          variant === 'primary',
        [`${defaultStyles} border-transparent`]: variant === 'ghost',
        [`${defaultStyles} border-red-800 text-red-800 dark:border-red-500 dark:text-red-500`]:
          variant === 'caution',
        'bg-red-700 hover:bg-red-600': variant === 'danger',
        'text-gray-50 px-5 border-transparent':
          variant === 'primary' || variant === 'danger',
        className: Boolean(className),
      })}
      {...props}
    >
      {children}
    </button>
  )
}
