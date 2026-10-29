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
        const tipContent =
          option.description ?? option.title ?? option.ariaLabel

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
              'focus:z-10',
              isSelected &&
                '!border-[color-mix(in_oklab,var(--primary-500)_70%,black)] !bg-[color-mix(in_oklab,var(--primary-500)_55%,black)] dark:!border-[color-mix(in_oklab,var(--primary-500)_40%,black)] dark:!bg-[color-mix(in_oklab,var(--primary-500)_35%,black)] !text-white !shadow-[inset_0_1px_2px_rgba(0,0,0,0.25)]',
              isFirst && '!rounded-e-none',
              isLast && '!rounded-s-none !border-l-0',
              !isFirst && !isLast && '!rounded-none !border-l-0'
            )}
          >
            {option.label}
          </Button>
        )
      })}
    </span>
  )
}
