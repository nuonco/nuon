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
  as?: ElementType
  family?: TTextFamily
  flex?: boolean
  level?: 1 | 2 | 3 | 4 | 5 | 6
  nowrap?: boolean
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
  as,
  className,
  children,
  family = 'sans',
  flex,
  level,
  nowrap,
  role,
  variant = 'body',
  theme = 'default',
  weight = 'normal',
  ...props
}: IText) => {
  const isHeading = role === 'heading' || (level && !role)
  let Element: ElementType = 'span'
  if (as) Element = as
  else if (isHeading && level) Element = `h${level}` as const
  else if (role === 'paragraph') Element = 'p'
  else if (role === 'code') Element = 'code'
  else if (role === 'time') Element = 'time'

  const extraTracking =
    family === 'mono' && ['h1', 'h2', 'h3'].includes(variant)
      ? headingMonoTracking
      : ''

  let display: string
  if (flex) {
    display = 'inline-flex items-center gap-1.5'
  } else if (Element === 'span' || Element === 'time') {
    display = 'inline'
  } else {
    display = 'block'
  }

  return (
    <Element
      aria-level={isHeading && level ? level : undefined}
      className={cn(
        display,
        FAMILY_CLASSES[family],
        VARIANT_CLASSES[variant],
        WEIGHT_CLASSES[weight],
        THEME_CLASSES[theme],
        extraTracking,
        nowrap ? 'text-nowrap' : 'text-wrap',
        className
      )}
      role={isHeading ? 'heading' : undefined}
      {...props}
    >
      {children}
    </Element>
  )
}
