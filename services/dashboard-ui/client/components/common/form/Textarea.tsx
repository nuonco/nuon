import { type TextareaHTMLAttributes, forwardRef, useState, useRef, useEffect } from 'react'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface ITextarea
  extends Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'size'> {
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
  autoResize?: boolean
  minRows?: number
  maxRows?: number
}

export const Textarea = forwardRef<HTMLTextAreaElement, ITextarea>(
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
      autoResize = false,
      minRows = 3,
      maxRows = 10,
      value,
      defaultValue,
      onChange,
      ...props
    },
    ref
  ) => {
    const [isInvalid, setIsInvalid] = useState(false)
    const [hasBlurred, setHasBlurred] = useState(false)
    const [showValidationMessage, setShowValidationMessage] = useState(false)
    const [internalValue, setInternalValue] = useState(defaultValue || '')
    const textareaRef = useRef<HTMLTextAreaElement>(null)

    const currentValue = value !== undefined ? value : internalValue

    // Auto-resize functionality
    useEffect(() => {
      if (autoResize && textareaRef.current) {
        const textarea = textareaRef.current
        
        const adjustHeight = () => {
          // Reset height to calculate scrollHeight properly
          textarea.style.height = 'auto'
          
          // Calculate line height and row constraints
          const lineHeight = parseInt(window.getComputedStyle(textarea).lineHeight)
          const minHeight = lineHeight * minRows
          const maxHeight = lineHeight * maxRows
          
          // Set height based on content, respecting min/max constraints
          const newHeight = Math.min(Math.max(textarea.scrollHeight, minHeight), maxHeight)
          textarea.style.height = `${newHeight}px`
          
          // Show scrollbar if content exceeds maxRows
          textarea.style.overflowY = textarea.scrollHeight > maxHeight ? 'auto' : 'hidden'
        }
        
        adjustHeight()
        
        // Adjust on content changes
        const handleInput = () => adjustHeight()
        textarea.addEventListener('input', handleInput)
        
        return () => textarea.removeEventListener('input', handleInput)
      }
    }, [autoResize, minRows, maxRows, currentValue])

    // Monitor validation state
    useEffect(() => {
      if (required && textareaRef.current) {
        const textarea = textareaRef.current
        
        const checkValidity = () => {
          // Only show invalid state after user blur or form submission attempt
          if (hasBlurred) {
            setIsInvalid(!textarea.checkValidity())
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
            if (textarea.checkValidity()) {
              setShowValidationMessage(false)
            }
          }
        }
        
        const handleBlur = () => {
          setHasBlurred(true)
          // Check validity after blur
          if (!textarea.checkValidity()) {
            setIsInvalid(true)
            setShowValidationMessage(true)
          }
        }
        
        textarea.addEventListener('invalid', handleInvalid)
        textarea.addEventListener('input', handleInput)
        textarea.addEventListener('blur', handleBlur)
        
        return () => {
          textarea.removeEventListener('invalid', handleInvalid)
          textarea.removeEventListener('input', handleInput)
          textarea.removeEventListener('blur', handleBlur)
        }
      }
    }, [required, hasBlurred])

    const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      if (value === undefined) {
        setInternalValue(e.target.value)
      }
      onChange?.(e)
    }

    const sizeClasses = {
      sm: 'px-2 py-1 text-sm',
      md: 'px-3 py-2 text-sm',
      lg: 'px-4 py-3 text-base',
    }

    const baseClasses = cn(
      // Base styles
      'w-full rounded-md border transition-colors duration-200 resize-vertical',
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
      
      // Auto-resize specific styles
      {
        'resize-none overflow-hidden': autoResize,
        'min-h-[4rem]': !autoResize, // Default min-height when not auto-resizing
      },
      
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

    const textarea = (
      <textarea
        ref={(node) => {
          textareaRef.current = node
          if (typeof ref === 'function') {
            ref(node)
          } else if (ref) {
            ref.current = node
          }
        }}
        className={baseClasses}
        disabled={disabled}
        required={required}
        value={currentValue}
        onChange={handleChange}
        rows={autoResize ? minRows : props.rows}
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
          {textarea}
          {renderDescription()}
        </div>
      )
    }

    return (
      <div className="space-y-1">
        {textarea}
        {renderDescription()}
      </div>
    )
  }
)

Textarea.displayName = 'Textarea'