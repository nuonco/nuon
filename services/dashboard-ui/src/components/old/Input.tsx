// @ts-ignore
import classNames from 'classnames'
import React, { type FC, forwardRef } from 'react'

export interface IRadioInput extends React.HTMLAttributes<HTMLInputElement> {
  checked?: boolean
  name: string
  labelClassName?: string
  labelText: React.ReactNode
  value: string
}

export const RadioInput: FC<IRadioInput> = ({
  className,
  labelClassName,
  labelText,
  ...props
}) => {
  return (
    <label
      className={classNames(
        'flex gap-3 items-center w-full px-4 py-2 cursor-pointer hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10',
        {
          [`${labelClassName}`]: Boolean(labelClassName),
        }
      )}
    >
      <input
        className={classNames('accent-primary-600 w-auto h-[14px]', {
          [`${className}`]: Boolean(className),
        })}
        {...props}
        type="radio"
      />
      <span className="font-medium text-xs">{labelText}</span>
    </label>
  )
}

export interface ICheckboxInput
  extends React.InputHTMLAttributes<HTMLInputElement> {
  checked?: boolean
  name: string
  labelText?: React.ReactNode
  labelClassName?: string
  labelTextClassName?: string
  ref?: any
  value?: string
}

export const CheckboxInput = forwardRef<HTMLInputElement, ICheckboxInput>(
  (
    { className, labelClassName, labelText, labelTextClassName, ...props },
    ref
  ) => {
    return (
      <label
        className={classNames(
          'flex gap-3 items-center w-full px-4 py-2 cursor-pointer hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10',
          {
            [`${labelClassName}`]: Boolean(labelClassName),
          }
        )}
      >
        <input
          className={classNames('accent-primary-600 w-auto h-[14px]', {
            [`${className}`]: Boolean(className),
          })}
          ref={ref}
          {...props}
          type="checkbox"
        />
        {labelText ? (
          <span
            className={classNames('font-medium text-xs', {
              [`${labelTextClassName}`]: Boolean(labelTextClassName),
            })}
          >
            {labelText}
          </span>
        ) : null}
      </label>
    )
  }
)

CheckboxInput.displayName = 'CheckboxInput'

export const Input: FC<
  React.InputHTMLAttributes<HTMLInputElement> & { isSearch?: boolean }
> = ({ className, isSearch = false, ...props }) => {
  return (
    <input
      className={classNames(
        'px-3 py-2 text-sm rounded border shadow-sm bg-cool-grey-50 dark:bg-dark-grey-800 [&:user-invalid]:border-red-600 [&:user-invalid]:dark:border-red-600 focus:outline outline-1 outline-primary-500 dark:outline-primary-400 disable-ligatures font-mono',
        {
          'bg-cool-grey-200 text-cool-grey-500 dark:bg-dark-grey-300 dark:text-dark-grey-900 cursor-not-allowed':
            props?.disabled,
          '!pl-8 !pr-3.5': isSearch,
          [`${className}`]: Boolean(className),
        }
      )}
      {...props}
    />
  )
}
