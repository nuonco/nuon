import classNames from 'classnames'
import React, { type FC } from 'react'

export interface IRadioInput extends React.HTMLAttributes<HTMLInputElement> {
  checked?: boolean
  name: string
  labelText: React.ReactNode
  value: string
}

export const RadioInput: FC<IRadioInput> = ({
  className,
  labelText,
  ...props
}) => {
  return (
    <label className="flex gap-3 items-center w-full px-4 py-2 cursor-pointer hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10">
      <input
        className={classNames('accent-primary-600 w-auto h-[14px]', {
          [`${className}`]: Boolean(className),
        })}
        {...props}
        type="radio"
      />
      <span className="font-medium text-sm">{labelText}</span>
    </label>
  )
}

export interface ICheckboxInput extends React.HTMLAttributes<HTMLInputElement> {
  checked?: boolean
  name: string
  labelText: React.ReactNode
  value: string
}

export const CheckboxInput: FC<ICheckboxInput> = ({
  className,
  labelText,
  ...props
}) => {
  return (
    <label className="flex gap-3 items-center w-full px-4 py-2 cursor-pointer hover:bg-black/5 focus:bg-black/5 active:bg-black/10 dark:hover:bg-white/5 dark:focus:bg-white/5 dark:active:bg-white/10">
      <input
        className={classNames('accent-primary-600 w-auto h-[14px]', {
          [`${className}`]: Boolean(className),
        })}
        {...props}
        type="checkbox"
      />
      <span className="font-medium text-sm">{labelText}</span>
    </label>
  )
}
