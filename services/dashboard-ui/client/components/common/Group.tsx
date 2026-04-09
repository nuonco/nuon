import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/utils/classnames'

type TGap = 1 | 1.5 | 2 | 3 | 4 | 6 | 8
type TAlign = 'start' | 'center' | 'end' | 'baseline' | 'stretch'
type TJustify = 'start' | 'center' | 'end' | 'between'

export interface IGroup extends HTMLAttributes<HTMLDivElement> {
  gap?: TGap
  align?: TAlign
  justify?: TJustify
  wrap?: boolean
  children: ReactNode
}

const GAP: Record<TGap, string> = {
  1: 'gap-1',
  1.5: 'gap-1.5',
  2: 'gap-2',
  3: 'gap-3',
  4: 'gap-4',
  6: 'gap-6',
  8: 'gap-8',
}

const ALIGN: Record<TAlign, string> = {
  start: 'items-start',
  center: 'items-center',
  end: 'items-end',
  baseline: 'items-baseline',
  stretch: 'items-stretch',
}

const JUSTIFY: Record<TJustify, string> = {
  start: 'justify-start',
  center: 'justify-center',
  end: 'justify-end',
  between: 'justify-between',
}

export const Group = ({
  gap = 4,
  align = 'center',
  justify = 'start',
  wrap = false,
  className,
  children,
  ...props
}: IGroup) => (
  <div
    className={cn(
      'flex',
      GAP[gap],
      ALIGN[align],
      JUSTIFY[justify],
      wrap && 'flex-wrap',
      className,
    )}
    {...props}
  >
    {children}
  </div>
)
