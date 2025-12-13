import type { ElementType, HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'

export type TTextFamily = 'sans' | 'mono'
export type TTextVariant =
  | 'h1'
  | 'h2'
  | 'h3'
  | 'base'
  | 'body'
  | 'subtext'
  | 'label'
export type TTextWeight = 'normal' | 'strong' | 'stronger'
export type TTextTheme =
  | 'default'
  | 'neutral'
  | 'info'
  | 'warn'
  | 'error'
  | 'success'
  | 'brand'

export interface IText extends HTMLAttributes<HTMLSpanElement> {
  family?: TTextFamily
  level?: 1 | 2 | 3 | 4 | 5 | 6
  role?: 'paragraph' | 'heading' | 'code' | 'time'
  theme?: TTextTheme
  variant?: TTextVariant
  weight?: TTextWeight
}

const FAMILY_CLASSES: Record<TTextFamily, string> = {
  sans: 'font-sans',
  mono: 'font-mono',
}

const VARIANT_CLASSES: Record<TTextVariant, string> = {
  h1: 'text-[34px] leading-10 tracking-[-0.8px]',
  h2: 'text-2xl leading-[30px] tracking-[-0.8px]',
  h3: 'text-lg leading-[27px] tracking-[-0.2px]',
  base: 'text-base leading-6 tracking-[-0.2px]',
  body: 'text-sm leading-6 tracking-[-0.2px]',
  subtext: 'text-xs leading-[17px] tracking-[-0.2px]',
  label: 'text-[11px] leading-[14px] tracking-[-0.2px]',
}

const WEIGHT_CLASSES: Record<TTextWeight, string> = {
  normal: 'font-normal',
  strong: 'font-medium',
  stronger: 'font-semibold',
}

const headingMonoTracking = 'tracking-[-0.2px]'

const THEME_CLASSES: Record<TTextTheme, string> = {
  default: '',
  neutral: 'text-white/70',
  info: 'text-blue-600',
  warn: 'text-orange-600',
  error: 'text-red-500',
  success: 'text-green-500',
  brand: 'text-primary-500',
}

export const Text = ({
  className,
  children,
  family = 'sans',
  level,
  role,
  variant = 'body',
  theme = 'default',
  weight = 'normal',
  ...props
}: IText) => {
  let Element: ElementType = 'span'
  if (role === 'heading' && level) Element = `h${level}` as const
  else if (role === 'paragraph') Element = 'p'
  else if (role === 'code') Element = 'code'
  else if (role === 'time') Element = 'time'

  const extraTracking =
    family === 'mono' && ['h1', 'h2', 'h3'].includes(variant)
      ? headingMonoTracking
      : ''

  return (
    <Element
      aria-level={role === 'heading' && level ? level : undefined}
      className={cn(
        Element === 'span' || Element === 'time' ? 'inline' : 'block',
        FAMILY_CLASSES[family],
        VARIANT_CLASSES[variant],
        WEIGHT_CLASSES[weight],
        THEME_CLASSES[theme],
        extraTracking,
        'text-wrap',
        className
      )}
      role={role === 'heading' ? 'heading' : undefined}
      {...props}
    >
      {children}
    </Element>
  )
}
