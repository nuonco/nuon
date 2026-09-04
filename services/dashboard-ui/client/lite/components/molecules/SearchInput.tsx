import { forwardRef, type KeyboardEvent, type ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Button } from '../atoms/Button'
import { Icon } from '../atoms/Icon'
import { Input, type IInput } from '../atoms/Input'
import { Kbd } from '../atoms/Kbd'
import { Text } from '../atoms/Text'

export interface ISearchInput
  extends Omit<IInput, 'type' | 'value' | 'defaultValue' | 'onChange'> {
  value: string
  onValueChange: (value: string) => void
  clearLabel?: string
  leading?: ReactNode
  inputClassName?: string
}

export const SearchInput = forwardRef<HTMLInputElement, ISearchInput>(
  (
    {
      value,
      onValueChange,
      clearLabel = 'Clear search',
      leading,
      inputClassName,
      className,
      disabled,
      loading,
      onKeyDown,
      ...props
    },
    ref
  ) => {
    const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
      if (event.key === 'Escape' && value) {
        event.preventDefault()
        event.stopPropagation()
        onValueChange('')
        return
      }
      onKeyDown?.(event)
    }

    return (
      <span className={cn('relative block min-w-0', className)}>
        <span className="pointer-events-none absolute top-1/2 left-2.5 flex -translate-y-1/2 text-tertiary">
          {leading ?? (
            <Icon variant="MagnifyingGlassIcon" size={14} aria-hidden />
          )}
        </span>
        <Input
          ref={ref}
          type="search"
          value={value}
          disabled={disabled}
          loading={loading}
          onChange={(event) => onValueChange(event.target.value)}
          onKeyDown={handleKeyDown}
          className={cn(
            'pr-8 pl-8 [&::-webkit-search-cancel-button]:appearance-none',
            inputClassName
          )}
          {...props}
        />
        {value && !disabled && !loading ? (
          <span className="absolute top-1/2 right-0.5 flex -translate-y-1/2">
            <Button
              size="sm"
              variant="ghost"
              iconOnly
              aria-label={clearLabel}
              tooltip={
                <span className="flex items-center gap-1.5">
                  <Text variant="caption">{clearLabel}</Text>
                  <Kbd>Esc</Kbd>
                </span>
              }
              onClick={() => onValueChange('')}
            >
              <Icon variant="XIcon" size={12} aria-hidden />
            </Button>
          </span>
        ) : null}
      </span>
    )
  }
)

SearchInput.displayName = 'SearchInput'
