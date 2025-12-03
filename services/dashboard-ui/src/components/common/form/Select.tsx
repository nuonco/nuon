import { type SelectHTMLAttributes, forwardRef } from 'react'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { cn } from '@/utils/classnames'

export interface SelectOption {
  value: string
  label: string
  disabled?: boolean
}

export interface ISelect
  extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  options: SelectOption[]
  labelProps?: Omit<ILabel, 'children'> & {
    labelText: string
    labelTextProps?: Omit<IText, 'children'>
  }
  helperText?: string
  helperTextProps?: Omit<IText, 'children'>
  error?: boolean
  errorMessage?: string
  errorMessageProps?: Omit<IText, 'children'>
  size?: 'sm' | 'md' | 'lg'
  placeholder?: string
}

export const Select = forwardRef<HTMLSelectElement, ISelect>(
  (
    {
      className,
      options,
      labelProps,
      helperText,
      helperTextProps = { variant: 'subtext' },
      error,
      errorMessage,
      errorMessageProps = { variant: 'subtext', theme: 'error' },
      size = 'md',
      disabled,
      placeholder,
      ...props
    },
    ref
  ) => {
    const sizeClasses = {
      sm: 'px-2 py-1 text-sm h-8',
      md: 'px-3 py-2 text-sm h-10',
      lg: 'px-4 py-3 text-base h-12',
    }

    const baseClasses = cn(
      'w-full rounded-md border transition-colors duration-200 appearance-none',
      'bg-white dark:bg-dark-grey-900',
      'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500',
      sizeClasses[size],
      {
        'border-cool-grey-300 dark:border-dark-grey-600': !error && !disabled,
        'text-cool-grey-900 dark:text-cool-grey-100': !disabled,
        'border-red-500 dark:border-red-400': error,
        'focus:ring-red-500 focus:border-red-500': error,
        'border-cool-grey-200 dark:border-dark-grey-700': disabled,
        'bg-cool-grey-50 dark:bg-dark-grey-800': disabled,
        'text-cool-grey-400 dark:text-cool-grey-500': disabled,
        'cursor-not-allowed': disabled,
      },
      className
    )

    const select = (
      <div className="relative">
        <select
          ref={ref}
          className={baseClasses}
          disabled={disabled}
          aria-invalid={error}
          aria-describedby={
            helperText || errorMessage ? `${props.id}-description` : undefined
          }
          {...props}
        >
          {placeholder && (
            <option value="" disabled>
              {placeholder}
            </option>
          )}
          {options.map((option) => (
            <option
              key={option.value}
              value={option.value}
              disabled={option.disabled}
            >
              {option.label}
            </option>
          ))}
        </select>
        <div className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
          <Icon
            variant="CaretDown"
            className={cn(
              'transition-colors',
              disabled
                ? 'text-cool-grey-400 dark:text-cool-grey-500'
                : 'text-cool-grey-500 dark:text-cool-grey-400'
            )}
          />
        </div>
      </div>
    )

    const renderDescription = () => {
      if (error && errorMessage) {
        return (
          <Text
            id={`${props.id}-description`}
            className={cn('block', errorMessageProps?.className)}
            {...errorMessageProps}
          >
            {errorMessage}
          </Text>
        )
      }

      if (helperText) {
        return (
          <Text
            id={`${props.id}-description`}
            className={cn('block', helperTextProps?.className)}
            {...helperTextProps}
          >
            {helperText}
          </Text>
        )
      }

      return null
    }

    if (labelProps) {
      return (
        <div className="flex flex-col gap-1">
          <Label
            className={cn('block', labelProps.className)}
            htmlFor={props.id}
            {...(labelProps as any)}
          >
            <Text
              className={cn('font-medium', labelProps.labelTextProps?.className)}
              variant="body"
              {...labelProps.labelTextProps}
            >
              {labelProps.labelText}
            </Text>
          </Label>
          {select}
          {renderDescription()}
        </div>
      )
    }

    return (
      <div className="flex flex-col gap-1">
        {select}
        {renderDescription()}
      </div>
    )
  }
)

Select.displayName = 'Select'