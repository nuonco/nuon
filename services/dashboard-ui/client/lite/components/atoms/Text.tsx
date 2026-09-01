import type { ElementType, HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'

export type TTextVariant =
  | 'display'
  | 'title'
  | 'heading'
  | 'body'
  | 'caption'
  | 'label'

export type TTextColor =
  | 'inherit'
  | 'primary'
  | 'secondary'
  | 'tertiary'
  | 'accent'
  | 'positive'

export type TTextWeight = 'normal' | 'medium' | 'semibold'

export type TTextFamily = 'sans' | 'mono'

export interface IText extends HTMLAttributes<HTMLElement> {
  as?: ElementType
  variant?: TTextVariant
  color?: TTextColor
  weight?: TTextWeight
  family?: TTextFamily
  lines?: number
  loading?: boolean
  loadingWidth?: number
}

const VARIANT_CLASSES: Record<TTextVariant, string> = {
  display: 'text-display',
  title: 'text-title',
  heading: 'text-heading',
  body: 'text-body',
  caption: 'text-caption',
  label: 'text-label',
}

const VARIANT_WEIGHTS: Record<TTextVariant, TTextWeight> = {
  display: 'semibold',
  title: 'semibold',
  heading: 'medium',
  body: 'normal',
  caption: 'normal',
  label: 'medium',
}

const COLOR_CLASSES: Record<TTextColor, string> = {
  inherit: '',
  primary: 'text-primary',
  secondary: 'text-secondary',
  tertiary: 'text-tertiary',
  accent: 'text-accent',
  positive: 'text-positive',
}

const WEIGHT_CLASSES: Record<TTextWeight, string> = {
  normal: 'font-normal',
  medium: 'font-medium',
  semibold: 'font-semibold',
}

const FAMILY_CLASSES: Record<TTextFamily, string> = {
  sans: 'font-sans',
  mono: 'font-mono',
}

const ZERO_WIDTH_SPACE = '\u200B'

const VARIANT_LOADING_WIDTH: Record<TTextVariant, number> = {
  display: 14,
  title: 12,
  heading: 10,
  body: 18,
  caption: 16,
  label: 8,
}

const LINE_CLAMP_CLASSES: Record<number, string> = {
  1: 'line-clamp-1',
  2: 'line-clamp-2',
  3: 'line-clamp-3',
  4: 'line-clamp-4',
  5: 'line-clamp-5',
  6: 'line-clamp-6',
}

export const Text = ({
  as: Element = 'span',
  variant = 'body',
  color = 'inherit',
  weight,
  family = 'sans',
  lines,
  loading,
  loadingWidth,
  className,
  children,
  ...props
}: IText) => {
  const typeClasses = cn(VARIANT_CLASSES[variant], FAMILY_CLASSES[family])

  if (loading) {
    const width = loadingWidth ?? VARIANT_LOADING_WIDTH[variant]
    const count = lines && lines > 1 ? lines : 1

    if (count === 1) {
      return (
        <Element
          aria-hidden
          className={cn('skeleton-text inline-block', typeClasses, className)}
          style={{ width: `${width}ch` }}
          {...props}
        >
          {ZERO_WIDTH_SPACE}
        </Element>
      )
    }

    return (
      <Element
        aria-hidden
        className={cn('flex flex-col', typeClasses, className)}
        {...props}
      >
        {Array.from({ length: count }, (_, index) => (
          <span
            key={index}
            className="skeleton-text block"
            style={{
              width: index === count - 1 ? `${Math.round(width * 0.6)}ch` : '100%',
            }}
          >
            {ZERO_WIDTH_SPACE}
          </span>
        ))}
      </Element>
    )
  }

  return (
    <Element
      className={cn(
        typeClasses,
        WEIGHT_CLASSES[weight ?? VARIANT_WEIGHTS[variant]],
        COLOR_CLASSES[color],
        lines ? LINE_CLAMP_CLASSES[lines] : undefined,
        className
      )}
      {...props}
    >
      {children}
    </Element>
  )
}
