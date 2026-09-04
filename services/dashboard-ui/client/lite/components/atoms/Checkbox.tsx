import {
  forwardRef,
  useEffect,
  useId,
  useImperativeHandle,
  useRef,
  type InputHTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/utils/classnames'
import { Icon } from './Icon'
import { Text } from './Text'

export interface ICheckbox
  extends Omit<
    InputHTMLAttributes<HTMLInputElement>,
    'type' | 'size' | 'required'
  > {
  label?: ReactNode
  description?: ReactNode
  error?: ReactNode
  indeterminate?: boolean
  loading?: boolean
}

export const Checkbox = forwardRef<HTMLInputElement, ICheckbox>(
  (
    {
      label,
      description,
      error,
      indeterminate = false,
      loading = false,
      className,
      disabled,
      id,
      ...props
    },
    forwardedRef
  ) => {
    const generatedId = useId()
    const inputId = id ?? `${generatedId}-checkbox`
    const descriptionId = description ? `${generatedId}-description` : undefined
    const errorId = error ? `${generatedId}-error` : undefined
    const ref = useRef<HTMLInputElement>(null)
    useImperativeHandle(forwardedRef, () => ref.current as HTMLInputElement)

    useEffect(() => {
      if (ref.current) ref.current.indeterminate = indeterminate
    }, [indeterminate])

    const control = (
      <span className="relative mt-0.5 flex size-5 shrink-0 items-center justify-center">
        <input
          ref={ref}
          id={inputId}
          type="checkbox"
          disabled={disabled || loading}
          aria-busy={loading || undefined}
          aria-invalid={!!error || undefined}
          aria-describedby={
            [descriptionId, errorId].filter(Boolean).join(' ') || undefined
          }
          className={cn(
            'peer absolute inset-0 size-5 cursor-pointer appearance-none rounded-md border border-field-border bg-field-bg outline-none transition-colors',
            'checked:border-action-primary checked:bg-action-primary indeterminate:border-action-primary indeterminate:bg-action-primary',
            'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring',
            'aria-invalid:border-field-invalid disabled:cursor-not-allowed disabled:opacity-50',
            loading && 'skeleton',
            className
          )}
          {...props}
        />
        {!loading ? (
          <>
            <span className="pointer-events-none relative hidden text-button-primary-text peer-checked:block peer-indeterminate:hidden">
              <Icon variant="CheckIcon" size={14} weight="bold" />
            </span>
            <span className="pointer-events-none relative hidden h-0.5 w-2.5 rounded-full bg-button-primary-text peer-indeterminate:block" />
          </>
        ) : null}
      </span>
    )

    if (!label && !description && !error) return control

    return (
      <label
        htmlFor={inputId}
        className={cn(
          'flex w-fit max-w-full gap-2.5 rounded-lg outline-none',
          disabled || loading
            ? 'cursor-not-allowed opacity-50'
            : 'cursor-pointer'
        )}
      >
        {control}
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
          {error ? (
            <Text id={errorId} variant="caption" className="text-field-invalid">
              {error}
            </Text>
          ) : null}
        </span>
      </label>
    )
  }
)

Checkbox.displayName = 'Checkbox'
