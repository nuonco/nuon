import { type InputHTMLAttributes, forwardRef, useState, useRef, useEffect } from 'react'
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
      ...props
    },
    ref
  ) => {
    const [isInvalid, setIsInvalid] = useState(false)
    const [hasBlurred, setHasBlurred] = useState(false)
    const [showValidationMessage, setShowValidationMessage] = useState(false)
    const inputRef = useRef<HTMLInputElement>(null)

    // Monitor validation state
    useEffect(() => {
      if (required && inputRef.current) {
        const input = inputRef.current
        
        const checkValidity = () => {
          // Only show invalid state after user blur or form submission attempt
          if (hasBlurred) {
            setIsInvalid(!input.checkValidity())
          }
        }
        
        // Check validity on value changes (only if already blurred)
        if (hasBlurred) {
          checkValidity()
        }
        
        // Listen for validation events (form submission attempts)
        const handleInvalid = (e: Event) => {
          e.preventDefault() // Prevent default browser validation message
          setHasBlurred(true)
          setIsInvalid(true)
          setShowValidationMessage(true)
        }
        
        const handleInput = () => {
          if (hasBlurred) {
            checkValidity()
            // Hide validation message when user enters valid input
            if (input.checkValidity()) {
              setShowValidationMessage(false)
            }
          }
        }
        
        const handleBlur = () => {
          setHasBlurred(true)
          // Check validity after blur
          if (!input.checkValidity()) {
            setIsInvalid(true)
            setShowValidationMessage(true)
          }
        }
        
        input.addEventListener('invalid', handleInvalid)
        input.addEventListener('input', handleInput)
        input.addEventListener('blur', handleBlur)
        
        return () => {
          input.removeEventListener('invalid', handleInvalid)
          input.removeEventListener('input', handleInput)
          input.removeEventListener('blur', handleBlur)
        }
      }
    }, [required, hasBlurred])

    const sizeClasses = {
      sm: 'px-2 py-1 text-sm h-8',
      md: 'px-3 py-2 text-sm h-10',
      lg: 'px-4 py-3 text-base h-12',
    }

    const baseClasses = cn(
      // Base styles
      'w-full rounded-md border transition-colors duration-200',
      'bg-white dark:bg-dark-grey-900',
      'placeholder:text-cool-grey-500 dark:placeholder:text-cool-grey-700',
      'font-mono',
      
      // Focus styles (brightest primary when focused)
      'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:!border-primary-500',
      
      // HTML5 validation states - user-invalid overrides everything
      'user-invalid:!border-red-500 user-invalid:dark:!border-red-400',
      'user-invalid:focus:!border-red-500 user-invalid:focus:!ring-red-500',
      
      // Size
      sizeClasses[size],
      
      // States
      {
        // Default state - dimmed primary (subtle but branded)
        '!border-primary-700 dark:!border-primary-400/50': !error && !disabled && !isInvalid,
        'text-cool-grey-900 dark:text-cool-grey-100': !disabled,
        
        // Error state - red overrides everything
        '!border-red-500 dark:!border-red-400': error || isInvalid,
        'focus:!ring-red-500 focus:!border-red-500': error || isInvalid,
        
        // Disabled state - grey overrides everything
        '!border-cool-grey-300 dark:!border-dark-grey-600': disabled,
        '!bg-cool-grey-100 dark:!bg-dark-grey-700': disabled,
        'text-cool-grey-400 dark:text-cool-grey-500': disabled,
        'cursor-not-allowed': disabled,
        'focus:!ring-transparent focus:!border-cool-grey-300 dark:focus:!border-dark-grey-600': disabled,
      },
      className
    )

    const input = (
      <input
        ref={(node) => {
          inputRef.current = node
          if (typeof ref === 'function') {
            ref(node)
          } else if (ref) {
            ref.current = node
          }
        }}
        className={baseClasses}
        disabled={disabled}
        required={required}
        aria-invalid={error || isInvalid}
        aria-describedby={
          helperText || errorMessage || showValidationMessage ? `${props.id}-description` : undefined
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