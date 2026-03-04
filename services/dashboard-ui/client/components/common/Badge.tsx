import type { HTMLAttributes } from 'react'
import type { TTheme } from '@/types'
import { cn } from '@/utils/classnames'

type TBadgeVariant = 'default' | 'code'
export type TBadgeTheme = TTheme

export interface IBadge extends HTMLAttributes<HTMLSpanElement> {
  size?: 'sm' | 'md' | 'lg'
  theme?: TBadgeTheme
  variant?: TBadgeVariant
}

const SIZE_CLASSES: Record<NonNullable<IBadge['size']>, string> = {
  sm: 'text-[11px] leading-[14px] tracking-[-0.2px] px-2 py-0.5',
  md: 'text-[12px] leading-[17px] tracking-[-0.2px] px-2 py-0.5',
  lg: 'text-[12px] leading-[17px] tracking-[-0.2px] px-3 py-1',
}

const VARIANT_CLASSES: Record<NonNullable<IBadge['variant']>, string> = {
  default: 'font-sans rounded-full',
  code: 'font-mono rounded-md',
}

const THEME_CLASSES: Record<NonNullable<IBadge['theme']>, string> = {
  brand:
    'bg-primary-50 !border-primary-200 text-primary-600 dark:bg-[#1B1026] dark:!border-[#351F4D] dark:text-primary-400',
  default:
    'bg-cool-grey-50 text-dark-grey-950 dark:bg-dark-grey-800 dark:text-cool-grey-100 !border-cool-grey-400 dark:!border-dark-grey-500',
  neutral:
    'bg-cool-grey-100 text-cool-grey-700 dark:bg-dark-grey-600 dark:text-cool-grey-500 !border-cool-grey-500 dark:!border-dark-grey-100',
  success:
    'bg-[#F4FBF7] text-green-800 !border-green-400 dark:bg-[#0C1B14] dark:!border-green-500/40 dark:text-green-500',
  warn: 'bg-[#FFF5EB] text-orange-800 !border-orange-400 dark:bg-[#2E1E10] dark:!border-orange-500/40 dark:text-orange-500',
  error:
    'bg-[#FEF2F2] text-red-800 !border-red-300 dark:bg-[#290C0D] dark:!border-red-500/40 dark:text-red-500',
  info: 'bg-[#FAFBFF] text-blue-800 !border-blue-400 dark:bg-[#0F172A] dark:!border-blue-500/40 dark:text-blue-500',
}

export const Badge = ({
  className,
  children,
  size = 'lg',
  theme = 'neutral',
  variant = 'default',
  ...props
}: IBadge) => {
  return (
    <span
      className={cn(
        'border flex gap-1.5 items-center shrink-0 grow-0 w-fit',
        SIZE_CLASSES[size],
        VARIANT_CLASSES[variant],
        THEME_CLASSES[theme],
        className
      )}
      {...props}
    >
      {children}
    </span>
  )
}
