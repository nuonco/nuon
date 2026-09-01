import type { ButtonHTMLAttributes, MouseEvent, ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import type { TPopoverSide } from '../../hooks/use-popover'
import { Spinner } from './Spinner'
import { Tooltip } from './Tooltip'

export type TButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'

export interface IButton extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: TButtonVariant
  loading?: boolean
  icon?: ReactNode
  iconOnly?: boolean
  tooltip?: ReactNode
  tooltipSide?: TPopoverSide
}

const BASE_CLASSES =
  'relative inline-grid grid-flow-col cursor-pointer items-center justify-center rounded-lg border text-body font-medium ' +
  'outline-none transition-colors duration-150 ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring ' +
  'aria-disabled:cursor-not-allowed aria-disabled:opacity-50'

const VARIANT_CLASSES: Record<TButtonVariant, string> = {
  primary:
    'border-transparent bg-button-primary text-button-primary-text ' +
    'not-aria-disabled:hover:bg-button-primary-hover not-aria-disabled:active:bg-button-primary-active',
  secondary:
    'border-button-secondary-border bg-button-secondary text-button-secondary-text ' +
    'not-aria-disabled:hover:bg-button-secondary-hover not-aria-disabled:active:bg-button-secondary-active',
  ghost:
    'border-transparent text-button-ghost-text ' +
    'not-aria-disabled:hover:bg-button-ghost-hover not-aria-disabled:hover:text-primary ' +
    'not-aria-disabled:active:bg-button-ghost-active',
  danger:
    'border-transparent bg-button-danger text-button-danger-text ' +
    'not-aria-disabled:hover:bg-button-danger-hover not-aria-disabled:active:bg-button-danger-active',
}

export const Button = ({
  variant = 'secondary',
  loading = false,
  icon,
  iconOnly = false,
  tooltip,
  tooltipSide = 'top',
  disabled,
  onClick,
  className,
  children,
  type = 'button',
  ...props
}: IButton) => {
  const inactive = disabled || loading

  const button = (
  <button
    type={type}
    aria-disabled={inactive || undefined}
    onClick={(event: MouseEvent<HTMLButtonElement>) => {
      if (inactive) {
        event.preventDefault()
        return
      }
      onClick?.(event)
    }}
    aria-busy={loading || undefined}
    className={cn(
      BASE_CLASSES,
      VARIANT_CLASSES[variant],
      iconOnly ? 'size-9' : 'h-9 px-3.5',
      className
    )}
    {...props}
  >
    <span
      aria-hidden={!loading}
      className={cn(
        'grid overflow-hidden transition-[grid-template-columns,margin-right] duration-200 ease-out motion-reduce:transition-none',
        loading ? 'grid-cols-[1fr]' : 'grid-cols-[0fr]',
        loading && !iconOnly && 'mr-1.5'
      )}
    >
      <span className="flex min-w-0 items-center justify-center">
        <Spinner />
      </span>
    </span>
    {!loading && icon && (
      <span className="-ml-0.5 mr-1.5 flex items-center">{icon}</span>
    )}
    {!(iconOnly && loading) && children}
  </button>
  )

  if (!tooltip) return button

  return (
    <Tooltip content={tooltip} side={tooltipSide}>
      {button}
    </Tooltip>
  )
}
