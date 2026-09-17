import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Button, type TButtonSize } from './Button'

export interface IToggleButtonOption<TValue extends string> {
  value: TValue
  label: ReactNode
  ariaLabel?: string
  tooltip?: ReactNode
  disabled?: boolean
  iconOnly?: boolean
}

export interface IToggleButton<TValue extends string>
  extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  value: TValue
  options: readonly [
    IToggleButtonOption<TValue>,
    IToggleButtonOption<TValue>,
    ...IToggleButtonOption<TValue>[],
  ]
  onValueChange: (value: NoInfer<TValue>) => void
  label: string
  size?: TButtonSize
}

export const ToggleButton = <const TValue extends string>({
  value,
  options,
  onValueChange,
  label,
  size = 'sm',
  className,
  ...props
}: IToggleButton<TValue>) => (
  <div
    role="group"
    aria-label={label}
    className={cn(
      'inline-flex items-center gap-0.5 rounded-lg border border-divider bg-field-bg p-0.5',
      className
    )}
    {...props}
  >
    {options.map((option) => {
      const selected = value === option.value

      return (
        <Button
          key={option.value}
          size={size}
          variant={selected ? 'secondary' : 'ghost'}
          iconOnly={option.iconOnly}
          aria-label={option.ariaLabel}
          aria-pressed={selected}
          tooltip={option.tooltip}
          disabled={option.disabled}
          onClick={() => onValueChange(option.value)}
        >
          {option.label}
        </Button>
      )
    })}
  </div>
)
