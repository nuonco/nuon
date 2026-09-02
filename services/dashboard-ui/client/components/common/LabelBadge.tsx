import type { CSSProperties, HTMLAttributes } from 'react'
import { useLayoutEffect, useRef, useState } from 'react'
import { Badge, type IBadge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { cn } from '@/utils/classnames'
import './LabelBadge.css'

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
  variant = 'code',
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

  const iconSize =
    size === 'lg' ? 13 : size === 'md' ? 12 : size === 'sm' ? 11 : 10

  const customStyle = customColor
    ? ({ '--label-color': customColor } as CSSProperties)
    : undefined

  const keyRef = useRef<HTMLSpanElement>(null)
  const valueRef = useRef<HTMLSpanElement>(null)
  const [truncated, setTruncated] = useState(false)

  useLayoutEffect(() => {
    const isOverflowing = (el: HTMLSpanElement | null) =>
      !!el && el.scrollWidth > el.clientWidth
    const check = () =>
      setTruncated(
        isOverflowing(keyRef.current) || isOverflowing(valueRef.current)
      )
    check()
    window.addEventListener('resize', check)
    return () => window.removeEventListener('resize', check)
  }, [key, value])

  const badge = (
    <span className={cn('inline-flex', className)} {...props}>
      <Badge
        size={size}
        theme={keyTheme}
        variant={variant}
        className="rounded-r-none"
      >
        <span ref={keyRef} className="block max-w-[14rem] truncate">
          {key}
        </span>
      </Badge>
      <Badge
        size={size}
        theme={customStyle ? 'none' : theme}
        variant={variant}
        className={cn(
          'rounded-l-none border-l-0',
          customStyle && 'label-badge-value',
          onRemove && 'pr-1'
        )}
        style={customStyle}
      >
        <span ref={valueRef} className="block max-w-[22rem] truncate">
          {value}
        </span>
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

  if (!truncated) return badge

  const fullLabel = value ? `${key}:${value}` : key

  return (
    <Tooltip
      tipContent={
        <Text
          variant="subtext"
          className="break-all"
          style={{ whiteSpace: 'normal' }}
        >
          {fullLabel}
        </Text>
      }
      tipContentClassName="max-w-sm"
    >
      {badge}
    </Tooltip>
  )
}
