import type { ReactNode } from 'react'
import { Handle, Position } from '@xyflow/react'
import { cn } from '@/utils/classnames'
import type { GraphAccent } from './accents'

export const NODE_WIDTH = 280
export const NODE_WIDTH_COMPACT = 160

interface IGroupNodeCard {
  accent: GraphAccent
  title: string
  headerRight?: ReactNode
  compact?: boolean
  children: ReactNode
}

export const GroupNodeCard = ({
  accent,
  title,
  headerRight,
  compact = false,
  children,
}: IGroupNodeCard) => (
  <>
    <Handle type="target" position={Position.Left} className="!opacity-0" />
    <div
      className={cn(
        'overflow-hidden rounded-lg border bg-white dark:bg-dark-grey-900',
        accent.border
      )}
      style={{ width: compact ? NODE_WIDTH_COMPACT : NODE_WIDTH }}
    >
      <div
        className={cn(
          'flex items-center justify-between gap-2 border-b',
          accent.border,
          accent.headerBg,
          compact ? 'px-2 py-1' : 'px-3 py-2'
        )}
      >
        <span
          className={cn(
            'truncate font-strong',
            accent.text,
            compact ? 'text-[10px]' : 'text-xs'
          )}
        >
          {title}
        </span>
        {headerRight}
      </div>
      <div className={cn('flex flex-col', compact ? 'gap-0.5 px-2 py-1' : 'gap-1 px-3 py-2')}>
        {children}
      </div>
    </div>
    <Handle type="source" position={Position.Right} className="!opacity-0" />
  </>
)
