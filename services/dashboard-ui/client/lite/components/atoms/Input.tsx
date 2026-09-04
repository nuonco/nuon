import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'

export type TFieldSize = 'sm' | 'md'

export interface IInput
  extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size' | 'required'> {
  size?: TFieldSize
  loading?: boolean
}

export const FIELD_CONTROL_CLASSES =
  'w-full rounded-lg border border-field-border bg-field-bg text-field-text outline-none transition-colors placeholder:text-field-placeholder not-disabled:hover:bg-field-bg-hover focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-focus-ring aria-invalid:border-field-invalid disabled:cursor-not-allowed disabled:opacity-50'

export const FIELD_SIZE_CLASSES: Record<TFieldSize, string> = {
  sm: 'h-7 px-2.5 text-caption',
  md: 'h-9 px-3 text-body',
}

export const Input = forwardRef<HTMLInputElement, IInput>(
  (
    {
      size = 'md',
      loading = false,
      disabled,
      className,
      value,
      defaultValue,
      ...props
    },
    ref
  ) => (
    <input
      ref={ref}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      value={value}
      defaultValue={defaultValue}
      className={cn(
        FIELD_CONTROL_CLASSES,
        FIELD_SIZE_CLASSES[size],
        loading && 'skeleton text-transparent placeholder:text-transparent',
        className
      )}
      {...props}
    />
  )
)

Input.displayName = 'Input'
