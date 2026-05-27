import type { ReactNode } from 'react'
import { Button, type TButtonSize } from '@/components/common/Button'
import { cn } from '@/utils/classnames'

export interface IToggleButtonOption<T extends string> {
  value: T
  label: ReactNode
  ariaLabel?: string
  title?: string
}

export interface IToggleButton<T extends string> {
  options: IToggleButtonOption<T>[]
  value: T
  onChange: (value: T) => void
  size?: TButtonSize
  className?: string
}

export const ToggleButton = <T extends string>({
  options,
  value,
  onChange,
  size = 'sm',
  className,
}: IToggleButton<T>) => {
  return (
    <span className={cn('flex items-center', className)}>
      {options.map((option, i) => {
        const isFirst = i === 0
        const isLast = i === options.length - 1

        return (
          <Button
            key={option.value}
            size={size}
            variant="secondary"
            isActive={value === option.value}
            onClick={() => onChange(option.value)}
            title={option.title}
            aria-label={option.ariaLabel}
            aria-pressed={value === option.value}
            className={cn(
              'focus:z-10',
              value !== option.value && '!bg-transparent !shadow-none !text-current',
              isFirst && '!rounded-e-none',
              isLast && '!rounded-s-none !border-l-0',
              !isFirst && !isLast && '!rounded-none !border-l-0',
            )}
          >
            {option.label}
          </Button>
        )
      })}
    </span>
  )
}
