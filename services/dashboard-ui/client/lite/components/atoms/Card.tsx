import type { ElementType, HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'

export type TCardPadding = 'none' | 'sm' | 'md' | 'lg'
export type TCardBlur = 'none' | 'sm' | 'md' | 'lg'

export interface ICard extends HTMLAttributes<HTMLDivElement> {
  as?: ElementType
  padding?: TCardPadding
  blur?: TCardBlur
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

export const Card = ({
  as: Component = 'div',
  padding = 'md',
  blur = 'md',
  interactive = false,
  className,
  children,
  ...props
}: ICard) => (
  <Component
    className={cn(
      'rounded-xl border border-card-border bg-card-bg shadow-[var(--card-shadow)]',
      PADDING_CLASSES[padding],
      BLUR_CLASSES[blur],
      interactive &&
        'cursor-pointer transition-colors hover:bg-card-bg-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring',
      className
    )}
    {...props}
  >
    {children}
  </Component>
)
