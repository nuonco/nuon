import { useMemo } from 'react'
import { Text } from '@/components/common/Text'
import type { TSpan } from '@/types'
import { cn } from '@/utils/classnames'
import {
  buildSpanForest,
  flattenForest,
  formatDurationNs,
  traceEnd,
  traceStart,
  type TSpanNode,
} from '@/utils/span-tree'

export interface ISpanTimeline {
  spans: TSpan[]
  selectedSpanId?: string
  onSelectSpan: (spanId: string) => void
}

const TICK_COUNT = 5

const buildTicks = (totalMs: number) =>
  Array.from({ length: TICK_COUNT + 1 }, (_, i) => ({
    pct: (i / TICK_COUNT) * 100,
    ms: (totalMs * i) / TICK_COUNT,
  }))

export const SpanTimeline = ({
  spans,
  selectedSpanId,
  onSelectSpan,
}: ISpanTimeline) => {
  const { rows, t0, totalMs } = useMemo(() => {
    const forest = buildSpanForest(spans)
    const flat = flattenForest(forest)
    const start = traceStart(spans)
    const end = traceEnd(spans)
    return {
      rows: flat,
      t0: start,
      totalMs: Math.max(end - start, 1),
    }
  }, [spans])

  const ticks = useMemo(() => buildTicks(totalMs), [totalMs])

  if (!spans?.length) {
    return (
      <div className="p-6 text-center">
        <Text variant="subtext" theme="neutral">
          No spans yet
        </Text>
      </div>
    )
  }

  return (
    <div className="relative flex flex-col py-1">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 px-2"
      >
        <div className="relative w-full h-full">
          {ticks.map((t, i) =>
            i === 0 || i === ticks.length - 1 ? null : (
              <div
                key={i}
                className="absolute top-0 bottom-0 w-px bg-cool-grey-200/60 dark:bg-dark-grey-700/60"
                style={{ left: `${t.pct}%` }}
              />
            )
          )}
        </div>
      </div>
      {rows.map((node) => (
        <SpanGanttRow
          key={node.span.span_id}
          node={node}
          t0={t0}
          totalMs={totalMs}
          isSelected={selectedSpanId === node.span.span_id}
          onSelect={onSelectSpan}
        />
      ))}
    </div>
  )
}

export interface ISpanTimelineAxis {
  spans: TSpan[]
  className?: string
}

export const SpanTimelineAxis = ({ spans, className }: ISpanTimelineAxis) => {
  const totalMs = useMemo(() => {
    if (!spans?.length) return 0
    return Math.max(traceEnd(spans) - traceStart(spans), 1)
  }, [spans])

  const ticks = useMemo(() => buildTicks(totalMs), [totalMs])

  if (!spans?.length) return null

  return (
    <div className={cn('relative h-full w-full', className)}>
      {ticks.map((t, i) => (
        <div
          key={i}
          className="absolute top-0 bottom-0 flex items-center"
          style={{
            left: `${t.pct}%`,
            transform:
              i === 0
                ? 'translateX(0)'
                : i === ticks.length - 1
                  ? 'translateX(-100%)'
                  : 'translateX(-50%)',
          }}
        >
          <Text variant="label" family="mono" theme="neutral" nowrap as="span">
            {formatDurationNs(t.ms * 1_000_000)}
          </Text>
        </div>
      ))}
    </div>
  )
}

interface ISpanGanttRow {
  node: TSpanNode
  t0: number
  totalMs: number
  isSelected: boolean
  onSelect: (spanId: string) => void
}

const SpanGanttRow = ({
  node,
  t0,
  totalMs,
  isSelected,
  onSelect,
}: ISpanGanttRow) => {
  const { span } = node
  const start = new Date(span.start_time).getTime() - t0
  const end = new Date(span.end_time).getTime() - t0
  const leftPct = clampPct((start / totalMs) * 100)
  const widthPct = Math.max(0.5, clampPct(((end - start) / totalMs) * 100))
  const isError = (span.status_code ?? '').toLowerCase() === 'error'

  return (
    <button
      type="button"
      onClick={() => onSelect(span.span_id)}
      title={`${span.name} · ${formatDurationNs(span.duration_ns)}`}
      className={cn(
        'group relative flex items-center min-h-7 px-2 text-left',
        'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-500',
        isSelected &&
          'bg-primary-200 dark:bg-primary-600/25'
      )}
    >
      <div className="relative h-3 w-full">
        <div
          className={cn(
            'absolute top-0 h-full rounded-sm',
            isError && 'bg-red-500',
            !isError && node.hasErrorDescendant && 'bg-red-400/70',
            !isError && !node.hasErrorDescendant && 'bg-emerald-500'
          )}
          style={{
            left: `${leftPct}%`,
            width: `${widthPct}%`,
            minWidth: '2px',
          }}
        />
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
