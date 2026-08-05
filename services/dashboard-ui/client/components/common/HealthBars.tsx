import type React from 'react'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { cn } from '@/utils/classnames'

const HEARTBEAT_WRAPPER_CLASS =
  'transition-all duration-fastest ease-cubic group heartbeat-item-parent [&:has(+.heartbeat-item-parent:hover)]:scale-y-[1.15] [&:hover+.heartbeat-item-parent_.heartbeat-item]:scale-y-[1.15]'
const HEARTBEAT_BAR_CLASS =
  'transition-all duration-fastest ease-cubic heartbeat-item group-hover:scale-y-[1.3]'

export interface IHealthBar {
  key: string | number
  colorClass: string
  ariaLabel?: string
  tooltip: React.ReactNode
}

export interface IHealthBars {
  bars: IHealthBar[]
  animated?: boolean
  grow?: boolean
  barClassName?: string
  emptyMessage?: string
  className?: string
}

export const HealthBars = ({
  bars,
  animated = false,
  grow = false,
  barClassName,
  emptyMessage,
  className,
}: IHealthBars) => {
  if (!bars.length) {
    if (!emptyMessage) return null
    return (
      <div className="flex items-center justify-center h-10 rounded-md border border-white/5 bg-white/[0.02] dark:border-white/5 dark:bg-white/[0.02]">
        <Text variant="subtext" theme="neutral">
          {emptyMessage}
        </Text>
      </div>
    )
  }

  return (
    <div className={cn('flex items-center gap-0.5', className)}>
      {bars.map((bar) => (
        <Tooltip
          key={bar.key}
          position="top"
          className={cn(grow && 'flex-auto', animated && HEARTBEAT_WRAPPER_CLASS)}
          tipContentClassName="!whitespace-normal !w-auto !p-2"
          tipContent={bar.tooltip}
        >
          <div
            aria-label={bar.ariaLabel}
            tabIndex={0}
            className={cn(
              'rounded-[2px] focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-1 focus-visible:outline-primary-400/80',
              grow ? 'w-full flex-auto' : 'w-1.5 shrink-0 grow',
              animated && HEARTBEAT_BAR_CLASS,
              barClassName,
              bar.colorClass
            )}
          />
        </Tooltip>
      ))}
    </div>
  )
}
