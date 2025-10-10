import type { ElementType, HTMLAttributes } from 'react'
import type { TTheme } from '@/types'
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
export type TTextTheme = TTheme

export interface IText extends HTMLAttributes<HTMLSpanElement> {
  family?: TTextFamily
  level?: 1 | 2 | 3 | 4 | 5 | 6
  role?: 'paragraph' | 'heading' | 'code' | 'time'
  theme?: TTextTheme
  variant?: TTextVariant
  weight?: TTextWeight
}

// Font family
const FAMILY_CLASSES: Record<TTextFamily, string> = {
  sans: 'font-sans', // should be mapped to your --font-inter in Tailwind config
  mono: 'font-mono', // should be mapped to your --font-hack in Tailwind config
}

// Font size, line height, letter spacing (tracking)
const VARIANT_CLASSES: Record<TTextVariant, string> = {
  h1: 'text-[34px] leading-10 tracking-[-0.8px]',
  h2: 'text-2xl leading-[30px] tracking-[-0.8px]',
  h3: 'text-lg leading-[27px] tracking-[-0.2px]',
  base: 'text-base leading-6 tracking-[-0.2px]',
  body: 'text-sm leading-6 tracking-[-0.2px]',
  subtext: 'text-xs leading-[17px] tracking-[-0.2px]',
  label: 'text-[11px] leading-[14px] tracking-[-0.2px]',
}

// Font weight
const WEIGHT_CLASSES: Record<TTextWeight, string> = {
  normal: 'font-normal',
  strong: 'font-strong',
  stronger: 'font-stronger',
}

// Special case: headings + mono = reduced letter spacing
// We'll add this with an extra class if matched
const headingMonoTracking = 'tracking-[-0.2px]'

// Theme colors
const THEME_CLASSES: Record<TTextTheme, string> = {
  default: '',
  neutral: 'text-cool-grey-600 dark:text-white/70',
  info: 'text-blue-800 dark:text-blue-600',
  warn: 'text-orange-800 dark:text-orange-600',
  error: 'text-red-800 dark:text-red-500',
  success: 'text-green-800 dark:text-green-500',
  brand: 'text-primary-600 dark:text-primary-500',
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
  // Optionally render as semantic element
  let Element: ElementType = 'span'
  if (role === 'heading' && level) Element = `h${level}` as const
  else if (role === 'paragraph') Element = 'p'
  else if (role === 'code') Element = 'code'
  else if (role === 'time') Element = 'time'

  // Add special tracking tweak for headings + mono
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
