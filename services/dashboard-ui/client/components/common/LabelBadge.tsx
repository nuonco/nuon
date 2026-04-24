import type { HTMLAttributes } from 'react'
import { Badge, type IBadge } from '@/components/common/Badge'
import { cn } from '@/utils/classnames'

export interface ILabelBadge extends HTMLAttributes<HTMLSpanElement> {
  label?: string
  labelKey?: string
  labelValue?: string
  keyTheme?: IBadge['theme']
  theme?: IBadge['theme']
  size?: IBadge['size']
  variant?: IBadge['variant']
}

export const LabelBadge = ({
  label,
  labelKey,
  labelValue,
  keyTheme = 'neutral',
  theme = 'info',
  size = 'lg',
  variant,
  className,
  ...props
}: ILabelBadge) => {
  let key = labelKey
  let value = labelValue

  if (label) {
    const colonIndex = label.indexOf(':')
    if (colonIndex !== -1) {
      key = label.slice(0, colonIndex)
      value = label.slice(colonIndex + 1)
    } else {
      key = label
      value = ''
    }
  }

  return (
    <span className={cn('inline-flex', className)} {...props}>
      <Badge size={size} theme={keyTheme} variant={variant} className="rounded-r-none">
        {key}
      </Badge>
      <Badge size={size} theme={theme} variant={variant} className="rounded-l-none border-l-0">
        {value}
      </Badge>
    </span>
  )
}
