import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'

interface IColorDot extends HTMLAttributes<HTMLSpanElement> {
  color?: string
  size?: number
}

export const ColorDot = ({
  color,
  size = 14,
  className,
  ...props
}: IColorDot) => (
  <span
    className={cn('inline-block rounded-sm border shrink-0', className)}
    style={{ backgroundColor: color, width: size, height: size }}
    {...props}
  />
)
