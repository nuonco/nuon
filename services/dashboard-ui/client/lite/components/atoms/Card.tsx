import { forwardRef, type ElementType, type HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'

export type TCardPadding = 'none' | 'sm' | 'md' | 'lg'
export type TCardBlur = 'none' | 'sm' | 'md' | 'lg'
export type TCardShadow = 'none' | 'default' | 'floating'
export type TCardOpacity = 'default' | 'strong' | 'solid'

export interface ICard extends HTMLAttributes<HTMLDivElement> {
  as?: ElementType
  padding?: TCardPadding
  blur?: TCardBlur
  opacity?: TCardOpacity
  shadow?: TCardShadow
  interactive?: boolean
}

const PADDING_CLASSES: Record<TCardPadding, string> = {
  none: '',
  sm: 'p-3',
  md: 'p-4',
  lg: 'p-6',
}

const BLUR_CLASSES: Record<TCardBlur, string> = {
  none: '',
  sm: 'backdrop-blur-sm',
  md: 'backdrop-blur-md',
  lg: 'backdrop-blur-xl',
}

const OPACITY_CLASSES: Record<TCardOpacity, string> = {
  default: 'bg-card-bg',
  strong: 'bg-card-bg-strong',
  solid: 'bg-card-bg-solid',
}

const INTERACTIVE_OPACITY_CLASSES: Record<TCardOpacity, string> = {
  default: 'hover:bg-card-bg-hover',
  strong: 'hover:bg-card-bg-strong-hover',
  solid: 'hover:bg-card-bg-solid-hover',
}

const SHADOW_CLASSES: Record<TCardShadow, string> = {
  none: 'shadow-none',
  default: 'shadow-[var(--card-shadow)]',
  floating: 'shadow-[var(--card-shadow-floating)]',
}

export const Card = forwardRef<HTMLDivElement, ICard>(
  (
    {
      as: Component = 'div',
      padding = 'md',
      blur = 'md',
      opacity = 'default',
      shadow = 'default',
      interactive = false,
      className,
      children,
      ...props
    },
    ref
  ) => (
    <Component
      ref={ref}
      className={cn(
        'rounded-xl border border-card-border',
        PADDING_CLASSES[padding],
        BLUR_CLASSES[blur],
        OPACITY_CLASSES[opacity],
        SHADOW_CLASSES[shadow],
        interactive &&
          'cursor-pointer transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring',
        interactive && INTERACTIVE_OPACITY_CLASSES[opacity],
        className
      )}
      {...props}
    >
      {children}
    </Component>
  )
)

Card.displayName = 'Card'
