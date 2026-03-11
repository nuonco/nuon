import { createPortal } from 'react-dom'
import { type SelectHTMLAttributes, forwardRef, useState, useRef, useEffect } from 'react'
import { Label, type ILabel } from '@/components/common/form/Label'
import { Text, type IText } from '@/components/common/Text'
import { Badge, type IBadge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { cn } from '@/utils/classnames'
import "./Select.css"

export interface SelectOption {
  value: string
  label: string
  disabled?: boolean
  badge?: {
    label: string
    theme?: IBadge['theme']
  }
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
    const [dropdownPosition, setDropdownPosition] = useState<{ top: number; left: number; width: number } | null>(null)
    const hiddenInputRef = useRef<HTMLInputElement>(null)
    const validationInputRef = useRef<HTMLInputElement>(null)
    const selectRef = useRef<HTMLDivElement>(null)
    const buttonRef = useRef<HTMLButtonElement>(null)
    const portalRef = useRef<HTMLDivElement>(null)

    const currentValue = value !== undefined
      ? options.find(option => option.value === value) || null
      : internalValue

    const sizeClasses = {
      sm: 'px-2 py-1 text-sm',
      md: 'px-3 py-2 text-sm',
      lg: 'px-4 py-3 text-base',
    }

    const handleToggle = () => {
      if (disabled) return
      if (!isOpen && buttonRef.current) {
        const rect = buttonRef.current.getBoundingClientRect()
        setDropdownPosition({
          top: rect.bottom + 4,
          left: rect.left,
          width: rect.width,
        })
      }
      setIsOpen(prev => !prev)
    }

    const closeDropdown = (wasOpen: boolean) => {
      setIsOpen(false)
      if (required && wasOpen) {
        setHasBlurred(true)
        if (validationInputRef.current && !validationInputRef.current.checkValidity()) {
          setIsInvalid(true)
          setShowValidationMessage(true)
        }
      }
    }

    useEffect(() => {
      const handleClickOutside = (event: MouseEvent) => {
        const target = event.target as Node
        const inSelect = selectRef.current?.contains(target)
        const inPortal = portalRef.current?.contains(target)
        if (!inSelect && !inPortal) {
          closeDropdown(isOpen)
        }
      }
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [required, isOpen])

    useEffect(() => {
      if (!isOpen) return
      const handleScroll = (e: Event) => {
        if (portalRef.current?.contains(e.target as Node)) return
        setIsOpen(false)
      }
      const handleResize = () => setIsOpen(false)
      window.addEventListener('scroll', handleScroll, true)
      window.addEventListener('resize', handleResize)
      return () => {
        window.removeEventListener('scroll', handleScroll, true)
        window.removeEventListener('resize', handleResize)
      }
    }, [isOpen])

    useEffect(() => {
      if (required && validationInputRef.current) {
        const input = validationInputRef.current

        const checkValidity = () => {
          if (hasBlurred) {
            setIsInvalid(!input.checkValidity())
          }
        }

        if (hasBlurred) {
          checkValidity()
        }

        const handleInvalid = (e: Event) => {
          e.preventDefault()
          setHasBlurred(true)
          setIsInvalid(true)
          setShowValidationMessage(true)
        }

        const handleInput = () => {
          if (hasBlurred) {
            checkValidity()
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

        {required && (
          <input
            ref={validationInputRef}
            type="text"
            value={currentValue?.value || ''}
            required
            style={{ position: 'absolute', left: '-9999px', opacity: 0, pointerEvents: 'none' }}
            tabIndex={-1}
            aria-hidden="true"
            onChange={() => {}}
          />
        )}

        <button
          ref={buttonRef}
          type="button"
          onClick={handleToggle}
          disabled={disabled}
          className={cn(
            'flex items-center justify-between w-full border border-solid rounded shadow-sm transition-all duration-300 font-mono',
            'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:!border-primary-500',
            'user-invalid:!border-red-500 user-invalid:dark:!border-red-400',
            'user-invalid:focus:!border-red-500 user-invalid:focus:!ring-red-500',
            sizeClasses[size],
            {
              '!bg-cool-grey-200 text-cool-grey-500 dark:!bg-dark-grey-600 dark:text-dark-grey-900 cursor-not-allowed': disabled,
              '!border-cool-grey-300 dark:!border-dark-grey-600': disabled,
              'focus:!ring-transparent focus:!border-cool-grey-300 dark:focus:!border-dark-grey-600': disabled,

              'bg-white dark:bg-dark-grey-900 text-cool-grey-900 dark:text-cool-grey-100': !disabled && !error && !isInvalid,
              '!border-primary-700 dark:!border-primary-400/50': !disabled && !error && !isInvalid,

              '!border-red-500 dark:!border-red-400': error || isInvalid,
              'focus:!ring-red-500 focus:!border-red-500': error || isInvalid,
            },
            className
          )}
        >
          <div className="flex items-center gap-2 flex-1 min-w-0">
            <span className={cn("truncate", { "text-cool-grey-500 dark:text-cool-grey-400": !currentValue })}>
              {currentValue?.label || placeholder || 'Select an option...'}
            </span>
            {currentValue?.badge && (
              <Badge size="sm" theme={currentValue.badge.theme}>
                {currentValue.badge.label}
              </Badge>
            )}
          </div>
          <Icon
            variant="CaretDown"
            className={cn(
              'ml-2 transition-transform flex-shrink-0',
              { 'rotate-180': isOpen }
            )}
          />
        </button>

        {required && showValidationMessage && isInvalid && !isOpen && (
          <Text variant="subtext" theme="error" className="mt-1">
            Please select an option
          </Text>
        )}
      </div>
    )

    const dropdownPortal = isOpen && dropdownPosition
      ? createPortal(
          <div
            ref={portalRef}
            style={{
              position: 'fixed',
              top: dropdownPosition.top,
              left: dropdownPosition.left,
              width: dropdownPosition.width,
              zIndex: 9999,
            }}
          >
            <TransitionDiv
              isVisible={isOpen}
              className="select-options bg-cool-grey-100 dark:bg-dark-grey-800 shadow-sm border rounded py-1 px-2 max-h-72 overflow-x-hidden overflow-y-auto"
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
                      'transition duration-200 px-2 py-1 -mx-1.5 cursor-pointer select-none rounded text-sm font-mono text-left flex items-center justify-between gap-2',
                      {
                        'text-white bg-primary-600': currentValue?.value === option.value,
                        'hover:bg-black/5 dark:hover:bg-white/5': currentValue?.value !== option.value && !option.disabled,
                        'opacity-50 cursor-not-allowed': option.disabled,
                      }
                    )}
                  >
                    <span className="truncate flex-1">{option.label}</span>
                    {option.badge && (
                      <Badge size="sm" theme={option.badge.theme}>
                        {option.badge.label}
                      </Badge>
                    )}
                  </button>
                ))}
              </div>
            </TransitionDiv>
          </div>,
          document.body
        )
      : null

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
      const { labelText, labelTextProps, ...restLabelProps } = labelProps
      return (
        <div className="flex flex-col gap-1">
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
          {selectComponent}
          {dropdownPortal}
          {renderDescription()}
        </div>
      )
    }

    return (
      <div className="flex flex-col gap-1">
        {selectComponent}
        {dropdownPortal}
        {renderDescription()}
      </div>
    )
  }
)

Select.displayName = 'Select'
