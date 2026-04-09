import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/utils/classnames'

type TGap = 1 | 1.5 | 2 | 3 | 4 | 6 | 8
type TAlign = 'start' | 'center' | 'end' | 'stretch'
type TJustify = 'start' | 'center' | 'end' | 'between'

export interface IStack extends HTMLAttributes<HTMLDivElement> {
  gap?: TGap
  align?: TAlign
  justify?: TJustify
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
  stretch: 'items-stretch',
}

const JUSTIFY: Record<TJustify, string> = {
  start: 'justify-start',
  center: 'justify-center',
  end: 'justify-end',
  between: 'justify-between',
}

export const Stack = ({
  gap = 4,
  align = 'stretch',
  justify = 'start',
  className,
  children,
  ...props
}: IStack) => (
  <div
    className={cn(
      'flex flex-col',
      GAP[gap],
      ALIGN[align],
      JUSTIFY[justify],
      className,
    )}
    {...props}
  >
    {children}
  </div>
)
