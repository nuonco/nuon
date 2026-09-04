import {
  forwardRef,
  useId,
  type ButtonHTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { Text } from './Text'

export interface ISwitch
  extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'onChange'> {
  checked: boolean
  onChange: (checked: boolean) => void
  label?: ReactNode
  description?: ReactNode
  error?: ReactNode
  loading?: boolean
}

export const Switch = forwardRef<HTMLButtonElement, ISwitch>(
  (
    {
      checked,
      onChange,
      label,
      description,
      error,
      loading = false,
      className,
      disabled,
      ...props
    },
    ref
  ) => {
    const generatedId = useId()
    const descriptionId = description ? `${generatedId}-description` : undefined
    const errorId = error ? `${generatedId}-error` : undefined

    return (
      <span className="flex w-fit max-w-full flex-col gap-0.5">
        <button
          ref={ref}
          type="button"
          role="switch"
          aria-checked={checked}
          aria-invalid={!!error || undefined}
          aria-describedby={
            [descriptionId, errorId].filter(Boolean).join(' ') || undefined
          }
          disabled={disabled || loading}
          aria-busy={loading || undefined}
          onClick={() => onChange(!checked)}
          className={cn(
            'group flex min-h-7 max-w-full items-start gap-2.5 rounded-lg text-left outline-none',
            'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring',
            disabled || loading
              ? 'cursor-not-allowed opacity-50'
              : 'cursor-pointer',
            className
          )}
          {...props}
        >
          <span
            aria-hidden
            className={cn(
              'relative mt-1 h-5 w-9 shrink-0 rounded-full border transition-colors',
              checked
                ? 'border-action-primary bg-action-primary'
                : 'border-field-border bg-field-bg group-hover:bg-field-bg-hover',
              error && 'border-field-invalid',
              loading && 'skeleton'
            )}
          >
            {!loading ? (
              <span
                className={cn(
                  'absolute top-0.5 size-3.5 rounded-full bg-field-text shadow-sm transition-transform',
                  checked ? 'translate-x-[17px]' : 'translate-x-0.5'
                )}
              />
            ) : null}
          </span>
          {label || description ? (
            <span className="flex min-w-0 flex-col gap-0.5">
              {label ? (
                <Text variant="body" color="secondary">
                  {label}
                </Text>
              ) : null}
              {description ? (
                <Text id={descriptionId} variant="caption" color="tertiary">
                  {description}
                </Text>
              ) : null}
            </span>
          ) : null}
        </button>
        {error ? (
          <Text
            id={errorId}
            variant="caption"
            className="pl-[46px] text-field-invalid"
          >
            {error}
          </Text>
        ) : null}
      </span>
    )
  }
)

Switch.displayName = 'Switch'
