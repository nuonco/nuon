import { DateTime } from 'luxon'
import { useMemo } from 'react'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TSpan } from '@/types'
import { cn } from '@/utils/classnames'
import {
  buildSpanForest,
  formatDurationNs,
  traceEnd,
  traceStart,
  type TSpanNode,
} from '@/utils/span-tree'

export interface ITimelineBars {
  spans: TSpan[]
  selectedSpanId?: string
  onSelectSpan: (spanId: string) => void
  collapsed: Set<string>
}

export const TimelineBars = ({
  spans,
  selectedSpanId,
  onSelectSpan,
  collapsed,
}: ITimelineBars) => {
  const { visibleNodes, t0, totalMs } = useMemo(() => {
    const forest = buildSpanForest(spans)
    const start = spans.length ? traceStart(spans) : 0
    const end = spans.length ? traceEnd(spans) : 1
    const total = Math.max(end - start, 1)
    const out: TSpanNode[] = []
    const walk = (n: TSpanNode) => {
      out.push(n)
      if (collapsed.has(n.span.span_id)) return
      for (const c of n.children) walk(c)
    }
    for (const r of forest) walk(r)
    return { visibleNodes: out, t0: start, totalMs: total }
  }, [spans, collapsed])

  const ticks = useMemo(() => computeTicks(totalMs), [totalMs])

  if (!spans?.length) return null

  return (
    <div className="flex flex-col">
      <TimeAxis ticks={ticks} totalMs={totalMs} />
      <div className="flex flex-col">
        {visibleNodes.map((node) => (
          <BarRow
            key={node.span.span_id}
            node={node}
            t0={t0}
            totalMs={totalMs}
            isSelected={selectedSpanId === node.span.span_id}
            onSelect={onSelectSpan}
          />
        ))}
      </div>
    </div>
  )
}

interface IBarRow {
  node: TSpanNode
  t0: number
  totalMs: number
  isSelected: boolean
  onSelect: (spanId: string) => void
}

const BarRow = ({ node, t0, totalMs, isSelected, onSelect }: IBarRow) => {
  const { span } = node
  const start = DateTime.fromISO(span.start_time).toMillis() - t0
  const end = DateTime.fromISO(span.end_time).toMillis() - t0
  const leftPct = clampPct((start / totalMs) * 100)
  const widthPct = Math.max(0.5, clampPct(((end - start) / totalMs) * 100))
  const isError = (span.status_code ?? '').toLowerCase() === 'error'
  const tipPosition = leftPct + widthPct > 75 ? 'left' as const : 'bottom' as const

  const tipContent = (
    <div className="flex flex-col gap-0.5">
      <div className="flex items-center gap-1.5">
        <Text variant="subtext" weight="strong">{span.name}</Text>
        {span.service_name && (
          <Text variant="label">· {span.service_name}</Text>
        )}
      </div>
      <Text variant="label">{formatDurationNs(span.duration_ns)}</Text>
      {isError && span.status_message && (
        <Text variant="label" theme="error" className="max-w-[16rem] truncate">
          {span.status_message}
        </Text>
      )}
    </div>
  )

  return (
    <button
      type="button"
      onClick={() => onSelect(span.span_id)}
      className={cn(
        'flex items-center min-h-7 px-2 py-1 cursor-pointer text-left',
        'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-500',
        isSelected && 'bg-primary-200 dark:bg-primary-600/25'
      )}
    >
      <div className="relative h-3 w-full">
        <Tooltip
          tipContent={tipContent}
          position={tipPosition}
          tipContentClassName="!whitespace-normal !w-auto"
          className="!absolute !top-0 !h-full"
          style={{
            left: `${leftPct}%`,
            width: `${widthPct}%`,
            minWidth: '2px',
          }}
        >
          <div
            className={cn(
              'h-full w-full rounded-sm',
              isError && 'bg-red-500',
              !isError && node.hasErrorDescendant && 'bg-red-400/70',
              !isError && !node.hasErrorDescendant && 'bg-emerald-500'
            )}
          />
        </Tooltip>
      </div>
    </button>
  )
}

const clampPct = (p: number) => {
  if (!Number.isFinite(p)) return 0
  if (p < 0) return 0
  if (p > 100) return 100
  return p
}

const formatTickMs = (ms: number): string => {
  if (ms < 1) return '0'
  return formatDurationNs(ms * 1_000_000)
}

const NICE_INTERVALS = [
  1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10_000, 15_000,
  30_000, 60_000, 120_000, 300_000, 600_000,
]

const computeTicks = (totalMs: number, targetCount = 5): { ms: number; pct: number }[] => {
  if (totalMs <= 0) return []
  const rawInterval = totalMs / targetCount
  const interval = NICE_INTERVALS.find((n) => n >= rawInterval) ?? rawInterval
  const ticks: { ms: number; pct: number }[] = [{ ms: 0, pct: 0 }]
  let t = interval
  while (t < totalMs) {
    ticks.push({ ms: t, pct: (t / totalMs) * 100 })
    t += interval
  }
  ticks.push({ ms: totalMs, pct: 100 })
  return ticks
}

const TimeAxis = ({ ticks }: { ticks: { ms: number; pct: number }[]; totalMs: number }) => (
  <div className="flex items-end h-14 border-b border-cool-grey-200 dark:border-dark-grey-600 px-2">
    <div className="relative w-full pb-1">
      {ticks.map((tick, i) => {
        const isFirst = i === 0
        const isLast = i === ticks.length - 1
        return (
          <span
            key={tick.ms}
            className="absolute bottom-0 text-[10px] leading-none text-cool-grey-500 dark:text-dark-grey-300 whitespace-nowrap"
            style={{
              left: `${tick.pct}%`,
              transform: isFirst
                ? 'none'
                : isLast
                  ? 'translateX(-100%)'
                  : 'translateX(-50%)',
            }}
          >
            {formatTickMs(tick.ms)}
          </span>
        )
      })}
    </div>
  </div>
)
