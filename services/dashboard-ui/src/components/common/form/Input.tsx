import { type InputHTMLAttributes, forwardRef } from 'react'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface IInput
  extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
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
}

export const Input = forwardRef<HTMLInputElement, IInput>(
  (
    {
      className,
      labelProps,
      helperText,
      helperTextProps = { variant: 'subtext' },
      error,
      errorMessage,
      errorMessageProps = { variant: 'subtext', theme: 'error' },
      size = 'md',
      disabled,
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
      // Base styles
      'w-full rounded-md border transition-colors duration-200',
      'bg-white dark:bg-dark-grey-900',
      'placeholder:text-cool-grey-500 dark:placeholder:text-cool-grey-400',
      
      // Focus styles
      'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500',
      
      // Size
      sizeClasses[size],
      
      // States
      {
        // Default state
        'border-cool-grey-300 dark:border-dark-grey-600': !error && !disabled,
        'text-cool-grey-900 dark:text-cool-grey-100': !disabled,
        
        // Error state
        'border-red-500 dark:border-red-400': error,
        'focus:ring-red-500 focus:border-red-500': error,
        
        // Disabled state
        'border-cool-grey-200 dark:border-dark-grey-700': disabled,
        'bg-cool-grey-50 dark:bg-dark-grey-800': disabled,
        'text-cool-grey-400 dark:text-cool-grey-500': disabled,
        'cursor-not-allowed': disabled,
      },
      className
    )

    const input = (
      <input
        ref={ref}
        className={baseClasses}
        disabled={disabled}
        aria-invalid={error}
        aria-describedby={
          helperText || errorMessage ? `${props.id}-description` : undefined
        }
        {...props}
      />
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
        <div className="space-y-1">
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
          {input}
          {renderDescription()}
        </div>
      )
    }

    return (
      <div className="space-y-1">
        {input}
        {renderDescription()}
      </div>
    )
  }
)

Input.displayName = 'Input'