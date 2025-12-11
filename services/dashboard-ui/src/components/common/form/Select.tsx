'use client'

import { type SelectHTMLAttributes, forwardRef, useState, useRef, useEffect } from 'react'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { Icon } from '@/components/common/Icon'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { cn } from '@/utils/classnames'
import "./Select.css"

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

export const Select = forwardRef<HTMLInputElement, ISelect>(
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
      defaultValue,
      value,
      onChange,
      name,
      required,
      ...props
    },
    ref
  ) => {
    const [isOpen, setIsOpen] = useState(false)
    const [internalValue, setInternalValue] = useState<SelectOption | null>(() => {
      const initialValue = value !== undefined ? value : defaultValue
      return options.find(option => option.value === initialValue) || null
    })
    const [isInvalid, setIsInvalid] = useState(false)
    const [hasBlurred, setHasBlurred] = useState(false)
    const [showValidationMessage, setShowValidationMessage] = useState(false)
    const hiddenInputRef = useRef<HTMLInputElement>(null)
    const validationInputRef = useRef<HTMLInputElement>(null)
    const selectRef = useRef<HTMLDivElement>(null)

    const currentValue = value !== undefined 
      ? options.find(option => option.value === value) || null 
      : internalValue

    const sizeClasses = {
      sm: 'px-2 py-1 text-sm',
      md: 'px-3 py-2 text-sm',
      lg: 'px-4 py-3 text-base',
    }

    // Close dropdown when clicking outside and handle blur
    useEffect(() => {
      const handleClickOutside = (event: MouseEvent) => {
        if (selectRef.current && !selectRef.current.contains(event.target as Node)) {
          const wasOpen = isOpen
          setIsOpen(false)
          // Only mark as blurred if the dropdown was actually open (user was interacting)
          if (required && wasOpen) {
            setHasBlurred(true)
            // Check validity after blur
            if (validationInputRef.current && !validationInputRef.current.checkValidity()) {
              setIsInvalid(true)
              setShowValidationMessage(true)
            }
          }
        }
      }

      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [required, isOpen])

    // Monitor validation state
    useEffect(() => {
      if (required && validationInputRef.current) {
        const input = validationInputRef.current
        
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
            // Hide validation message when user makes a selection
            if (input.checkValidity()) {
              setShowValidationMessage(false)
            }
          }
        }
        
        input.addEventListener('invalid', handleInvalid)
        input.addEventListener('input', handleInput)
        
        return () => {
          input.removeEventListener('invalid', handleInvalid)
          input.removeEventListener('input', handleInput)
        }
      }
    }, [required, currentValue, hasBlurred])

    const handleOptionSelect = (option: SelectOption) => {
      if (value === undefined) {
        setInternalValue(option)
      }
      
      if (hiddenInputRef.current) {
        hiddenInputRef.current.value = option.value
        const event = new Event('change', { bubbles: true })
        hiddenInputRef.current.dispatchEvent(event)
      }
      
      if (onChange) {
        const syntheticEvent = {
          target: { value: option.value, name },
          currentTarget: { value: option.value, name },
        } as React.ChangeEvent<HTMLSelectElement>
        
        onChange(syntheticEvent)
      }

      // Hide validation message when a selection is made
      setShowValidationMessage(false)
      setIsOpen(false)
    }

    const selectComponent = (
      <div className="relative select" ref={selectRef}>
        <input
          ref={hiddenInputRef}
          type="hidden"
          name={name}
          value={currentValue?.value || ''}
          required={required}
          {...(ref && typeof ref === 'function' ? {} : { ref })}
        />
        
        {/* Additional validation input for required selects */}
        {required && (
          <input
            ref={validationInputRef}
            type="text"
            value={currentValue?.value || ''}
            required
            style={{ position: 'absolute', left: '-9999px', opacity: 0, pointerEvents: 'none' }}
            tabIndex={-1}
            aria-hidden="true"
            onChange={() => {}} // Controlled input
          />
        )}
        
        <button
          type="button"
          onClick={() => {
            if (!disabled) {
              setIsOpen(!isOpen)
            }
          }}
          disabled={disabled}
          className={cn(
            'flex items-center justify-between w-full border border-solid rounded shadow-sm transition-all duration-300 font-mono',
            // Focus styles (brightest primary when focused)
            'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:!border-primary-500',
            // HTML5 validation states - user-invalid overrides everything
            'user-invalid:!border-red-500 user-invalid:dark:!border-red-400',
            'user-invalid:focus:!border-red-500 user-invalid:focus:!ring-red-500',
            sizeClasses[size],
            {
              // Disabled state - grey overrides everything
              '!bg-cool-grey-200 text-cool-grey-500 dark:!bg-dark-grey-600 dark:text-dark-grey-900 cursor-not-allowed': disabled,
              '!border-cool-grey-300 dark:!border-dark-grey-600': disabled,
              'focus:!ring-transparent focus:!border-cool-grey-300 dark:focus:!border-dark-grey-600': disabled,
              
              // Default state - dimmed primary (subtle but branded)
              'bg-white dark:bg-dark-grey-900 text-cool-grey-900 dark:text-cool-grey-100': !disabled && !error && !isInvalid,
              '!border-primary-700 dark:!border-primary-400/50': !disabled && !error && !isInvalid,
              
              // Error state - red overrides everything
              '!border-red-500 dark:!border-red-400': error || isInvalid,
              'focus:!ring-red-500 focus:!border-red-500': error || isInvalid,
            },
            className
          )}
        >
          <span className={cn("truncate", { "text-cool-grey-500 dark:text-cool-grey-400": !currentValue })}>
            {currentValue?.label || placeholder || 'Select an option...'}
          </span>
          <Icon 
            variant="CaretDown" 
            className={cn(
              'ml-2 transition-transform',
              { 'rotate-180': isOpen }
            )} 
          />
        </button>

        <TransitionDiv
          isVisible={isOpen}
          className="select-options absolute z-30 w-full bg-cool-grey-100 dark:bg-dark-grey-800 shadow-sm border rounded py-1 px-2 mt-1.5 max-h-72 overflow-x-hidden overflow-y-auto"
        >
          <div className="flex flex-col gap-1">
            {options.length === 0 && <div className="px-2 py-1 text-sm">No options available</div>}
            {options.map((option) => (
              <button
                key={option.value}
                type="button"
                onClick={() => handleOptionSelect(option)}
                disabled={option.disabled}
                className={cn(
                  'transition duration-200 px-2 py-1 -mx-1.5 cursor-pointer select-none truncate rounded text-sm font-mono text-left',
                  {
                    'text-white bg-primary-600': currentValue?.value === option.value,
                    'hover:bg-black/5 dark:hover:bg-white/5': currentValue?.value !== option.value && !option.disabled,
                    'opacity-50 cursor-not-allowed': option.disabled,
                  }
                )}
              >
                {option.label}
              </button>
            ))}
          </div>
        </TransitionDiv>
        
        {/* Custom validation message */}
        {required && showValidationMessage && isInvalid && !isOpen && (
          <Text variant="subtext" theme="error" className="mt-1">
            Please select an option
          </Text>
        )}
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
          {selectComponent}
          {renderDescription()}
        </div>
      )
    }

    return (
      <div className="flex flex-col gap-1">
        {selectComponent}
        {renderDescription()}
      </div>
    )
  }
)

Select.displayName = 'Select'
