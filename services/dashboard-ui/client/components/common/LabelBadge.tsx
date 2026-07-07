import type { CSSProperties, HTMLAttributes } from 'react'
import { Badge, type IBadge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { cn } from '@/utils/classnames'

export interface ILabelBadge extends HTMLAttributes<HTMLSpanElement> {
  label?: string
  labelKey?: string
  labelValue?: string
  keyTheme?: IBadge['theme']
  theme?: IBadge['theme']
  size?: IBadge['size']
  variant?: IBadge['variant']
  onRemove?: () => void
  removeAriaLabel?: string
  disabled?: boolean
  customColor?: string
}

export const LabelBadge = ({
  label,
  labelKey,
  labelValue,
  keyTheme = 'neutral',
  theme = 'info',
  size = 'lg',
  variant,
  customColor,
  className,
  onRemove,
  removeAriaLabel = 'Remove label',
  disabled,
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

  const iconSize = size === 'lg' ? 13 : size === 'md' ? 12 : 11

  const customStyle: CSSProperties | undefined = customColor
    ? {
        backgroundColor: `${customColor}15`,
        color: customColor,
        borderColor: `${customColor}40`,
      }
    : undefined

  return (
    <span className={cn('inline-flex', className)} {...props}>
      <Badge size={size} theme={keyTheme} variant={variant} className="rounded-r-none">
        {key}
      </Badge>
      <Badge
        size={size}
        theme={customStyle ? undefined : theme}
        variant={variant}
        className={cn('rounded-l-none border-l-0', customStyle && 'border', onRemove && 'pr-1')}
        style={customStyle}
      >
        {value}
        {onRemove && (
          <button
            type="button"
            onClick={onRemove}
            disabled={disabled}
            aria-label={removeAriaLabel}
            className="ml-0.5 inline-flex items-center hover:text-red-600 disabled:opacity-50"
          >
            <Icon variant="XIcon" size={iconSize} />
          </button>
        )}
      </Badge>
    </span>
  )
}
