import {
  type InputHTMLAttributes,
  type ChangeEvent,
  type FocusEvent,
  type FormEvent,
  forwardRef,
  useState,
} from 'react'
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
      required,
      onChange,
      onBlur,
      onInvalid,
      ...props
    },
    ref
  ) => {
    const [isInvalid, setIsInvalid] = useState(false)
    const [hasBlurred, setHasBlurred] = useState(false)
    const [showValidationMessage, setShowValidationMessage] = useState(false)

    const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
      if (required && hasBlurred) {
        const valid = e.currentTarget.checkValidity()
        setIsInvalid(!valid)
        if (valid) {
          setShowValidationMessage(false)
        }
      }
      onChange?.(e)
    }

    const handleBlur = (e: FocusEvent<HTMLInputElement>) => {
      setHasBlurred(true)
      if (required && !e.currentTarget.checkValidity()) {
        setIsInvalid(true)
        setShowValidationMessage(true)
      }
      onBlur?.(e)
    }

    const handleInvalid = (e: FormEvent<HTMLInputElement>) => {
      e.preventDefault()
      setHasBlurred(true)
      setIsInvalid(true)
      setShowValidationMessage(true)
      onInvalid?.(e)
    }

    const sizeClasses = {
      sm: 'px-2 py-1 text-sm h-8',
      md: 'px-3 py-2 text-sm h-9',
      lg: 'px-4 py-3 text-base h-12',
    }

    const baseClasses = cn(
      'w-full rounded-md border transition-colors duration-200',
      'bg-white dark:bg-dark-grey-900',
      'shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)]',
      'placeholder:text-cool-grey-500 dark:placeholder:text-cool-grey-600',
      'font-sans',

      'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:!border-primary-500',

      'user-invalid:!border-red-500 user-invalid:dark:!border-red-400',
      'user-invalid:focus:!border-red-500 user-invalid:focus:!ring-red-500',

      sizeClasses[size],

      {
        'border-cool-grey-500/24 dark:border-cool-grey-500/24': !error && !disabled && !isInvalid,
        'text-cool-grey-900 dark:text-cool-grey-100': !disabled,

        '!border-red-500 dark:!border-red-400': error || isInvalid,
        'focus:!ring-red-500 focus:!border-red-500': error || isInvalid,

        '!border-cool-grey-300 dark:!border-dark-grey-600': disabled,
        '!bg-cool-grey-100 dark:!bg-dark-grey-700': disabled,
        'text-cool-grey-400 dark:text-cool-grey-500': disabled,
        'cursor-not-allowed': disabled,
        '!shadow-none': disabled,
        'focus:!ring-transparent focus:!border-cool-grey-300 dark:focus:!border-dark-grey-600': disabled,
      },
      className
    )

    const input = (
      <input
        ref={ref}
        className={baseClasses}
        disabled={disabled}
        required={required}
        aria-invalid={error || isInvalid}
        aria-describedby={
          helperText || errorMessage || showValidationMessage ? `${props.id}-description` : undefined
        }
        onChange={handleChange}
        onBlur={handleBlur}
        onInvalid={handleInvalid}
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

      if (required && showValidationMessage && isInvalid) {
        return (
          <Text
            id={`${props.id}-description`}
            variant="subtext"
            theme="error"
            className="mt-1"
          >
            Please fill out this field
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
      const { labelText, labelTextProps, ...restLabelProps } = labelProps
      return (
        <div className="space-y-1">
          <Label
            className={cn('block', labelProps.className)}
            htmlFor={props.id}
            {...restLabelProps}
          >
            <Text
              className={cn('font-medium', labelTextProps?.className)}
              variant="body"
              {...labelTextProps}
            >
              {labelText}
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