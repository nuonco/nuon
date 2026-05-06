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
    <div className="flex flex-col gap-1 p-2">
      <div className="grid grid-cols-[14rem_1fr] gap-2 px-2 py-1 border-b">
        <Text variant="subtext" theme="neutral" weight="strong">
          Span
        </Text>
        <div className="relative">
          <Text variant="subtext" theme="neutral" weight="strong">
            Timeline ({formatDurationNs(totalMs * 1_000_000)})
          </Text>
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
      className={cn(
        'grid grid-cols-[14rem_1fr] gap-2 items-center px-2 py-1 rounded text-left',
        'hover:bg-black/5 dark:hover:bg-white/5',
        isSelected && '!bg-primary-600/20 dark:!bg-primary-400/20'
      )}
    >
      <div
        className="flex items-center min-w-0"
        style={{ paddingLeft: `${node.depth * 12}px` }}
        title={span.name}
      >
        <Text family="mono" variant="subtext" nowrap as="span" className="truncate">
          {span.name}
        </Text>
      </div>
      <div className="relative h-4 w-full bg-cool-grey-100 dark:bg-dark-grey-800 rounded">
        <div
          className={cn(
            'absolute top-0 h-full rounded',
            isError && 'bg-red-500',
            !isError && node.hasErrorDescendant && 'bg-red-400/70',
            !isError && !node.hasErrorDescendant && 'bg-emerald-500'
          )}
          style={{ left: `${leftPct}%`, width: `${widthPct}%` }}
          title={`${span.name} · ${formatDurationNs(span.duration_ns)}`}
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
