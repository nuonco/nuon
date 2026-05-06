import { useMemo } from 'react'
import type { TSpan } from '@/types'
import { cn } from '@/utils/classnames'
import {
  buildSpanForest,
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

  if (!spans?.length) return null

  return (
    <div className="flex flex-col">
      <div className="h-14 border-b border-cool-grey-200 dark:border-dark-grey-600" />
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
  const start = new Date(span.start_time).getTime() - t0
  const end = new Date(span.end_time).getTime() - t0
  const leftPct = clampPct((start / totalMs) * 100)
  const widthPct = Math.max(0.5, clampPct(((end - start) / totalMs) * 100))
  const isError = (span.status_code ?? '').toLowerCase() === 'error'

  return (
    <button
      type="button"
      onClick={() => onSelect(span.span_id)}
      title={span.name}
      className={cn(
        'flex items-center min-h-7 px-2 py-1 cursor-pointer text-left',
        'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-500',
        isSelected && 'bg-primary-200 dark:bg-primary-600/25'
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
