import type { ButtonHTMLAttributes } from 'react'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

export interface IToggle
  extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'onChange'> {
  checked: boolean
  onChange: (checked: boolean) => void
  label?: string
  description?: string
}

export const Toggle = ({
  checked,
  onChange,
  label,
  description,
  className,
  disabled,
  ...props
}: IToggle) => {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      className={cn(
        'flex gap-2 text-left items-center',
        disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer',
        className
      )}
      onClick={() => onChange(!checked)}
      {...props}
    >
      <span className="relative shrink-0 h-[20px] w-[24px]">
        <span
          className={cn(
            'absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2',
            'h-[14px] w-[24px] rounded-full overflow-hidden transition-colors duration-200',
            checked
              ? 'bg-primary-600 dark:bg-primary-500'
              : 'bg-neutral-200 dark:bg-neutral-600'
          )}
        >
          <span
            className={cn(
              'absolute top-1/2 -translate-y-1/2 h-[10px] w-[10px] rounded-full bg-white',
              'shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)] transition-all duration-200',
              checked ? 'right-[2px]' : 'right-[12px]'
            )}
          />
        </span>
      </span>
      {(label || description) && (
        <span className="flex flex-col gap-[2px]">
          {label && (
            <Text variant="body" weight="strong">
              {label}
            </Text>
          )}
          {description && (
            <Text variant="subtext" theme="neutral">
              {description}
            </Text>
          )}
        </span>
      )}
    </button>
  )
}
