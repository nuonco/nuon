import type { ReactNode } from 'react'
import { Button, type TButtonSize } from '@/components/common/Button'
import { cn } from '@/utils/classnames'

export interface IToggleButtonOption<T extends string> {
  value: T
  label: ReactNode
  ariaLabel?: string
  title?: string
  description?: ReactNode
}

export interface IToggleButton<T extends string> {
  options: IToggleButtonOption<T>[]
  value: T
  onChange: (value: T) => void
  size?: TButtonSize
  className?: string
  label?: string
}

export const ToggleButton = <T extends string>({
  options,
  value,
  onChange,
  size = 'sm',
  className,
  label,
}: IToggleButton<T>) => {
  return (
    <span
      role="group"
      aria-label={label}
      className={cn('flex items-center', className)}
    >
      {options.map((option, i) => {
        const isFirst = i === 0
        const isLast = i === options.length - 1
        const isSelected = value === option.value
        const tipContent = option.description ?? option.title ?? option.ariaLabel

        return (
          <Button
            key={option.value}
            size={size}
            variant="secondary"
            isActive={isSelected}
            onClick={() => onChange(option.value)}
            aria-label={option.ariaLabel}
            aria-pressed={isSelected}
            tooltipProps={
              tipContent
                ? {
                    position: 'bottom',
                    tipContent,
                    tipContentClassName: 'max-w-72 whitespace-normal',
                    className: 'flex',
                  }
                : undefined
            }
            className={cn(
              'focus:z-10 !shadow-none',
              isSelected
                ? '!bg-primary-200 dark:!bg-primary-600/25 !text-primary-800 dark:!text-primary-400 !font-stronger'
                : '!bg-transparent !text-cool-grey-800 dark:!text-cool-grey-400 hover:!bg-cool-grey-500/8 dark:hover:!bg-cool-grey-500/8',
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
