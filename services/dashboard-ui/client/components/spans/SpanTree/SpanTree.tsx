import { useMemo, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TSpan } from '@/types'
import { cn } from '@/utils/classnames'
import { buildSpanForest, formatDurationNs, type TSpanNode } from '@/utils/span-tree'

export interface ISpanTree {
  spans: TSpan[]
  selectedSpanId?: string
  onSelectSpan: (spanId: string) => void
}

export const SpanTree = ({ spans, selectedSpanId, onSelectSpan }: ISpanTree) => {
  const forest = useMemo(() => buildSpanForest(spans), [spans])
  const allIds = useMemo(() => collectIds(forest), [forest])
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())

  const toggle = (id: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const expandAll = () => setCollapsed(new Set())
  const collapseAll = () => setCollapsed(new Set(allIds))

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
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2 px-2 py-1 border-b">
        <Button size="xs" variant="ghost" onClick={expandAll}>
          Expand all
        </Button>
        <Button size="xs" variant="ghost" onClick={collapseAll}>
          Collapse all
        </Button>
      </div>
      <div className="flex flex-col">
        {forest.map((node) => (
          <SpanTreeNode
            key={node.span.span_id}
            node={node}
            collapsed={collapsed}
            onToggle={toggle}
            selectedSpanId={selectedSpanId}
            onSelectSpan={onSelectSpan}
          />
        ))}
      </div>
    </div>
  )
}

const collectIds = (forest: TSpanNode[]): string[] => {
  const out: string[] = []
  const walk = (n: TSpanNode) => {
    out.push(n.span.span_id)
    for (const c of n.children) walk(c)
  }
  for (const r of forest) walk(r)
  return out
}

interface ISpanTreeNode {
  node: TSpanNode
  collapsed: Set<string>
  onToggle: (spanId: string) => void
  selectedSpanId?: string
  onSelectSpan: (spanId: string) => void
}

const SpanTreeNode = ({
  node,
  collapsed,
  onToggle,
  selectedSpanId,
  onSelectSpan,
}: ISpanTreeNode) => {
  const { span } = node
  const hasChildren = node.children.length > 0
  const isCollapsed = collapsed.has(span.span_id)
  const isSelected = selectedSpanId === span.span_id
  const status = statusFor(node)

  return (
    <>
      <div
        role="treeitem"
        aria-selected={isSelected}
        aria-expanded={hasChildren ? !isCollapsed : undefined}
        className={cn(
          'group grid grid-cols-[1.25rem_1rem_1fr_auto] items-center gap-2 py-1 px-2 rounded cursor-pointer',
          'hover:bg-black/5 dark:hover:bg-white/5',
          isSelected && '!bg-primary-600/20 dark:!bg-primary-400/20'
        )}
        style={{ paddingLeft: `${node.depth * 16 + 8}px` }}
        onClick={() => onSelectSpan(span.span_id)}
        title={span.name}
      >
        {hasChildren ? (
          <button
            type="button"
            className="flex items-center justify-center w-5 h-5 text-cool-grey-500"
            onClick={(e) => {
              e.stopPropagation()
              onToggle(span.span_id)
            }}
            aria-label={isCollapsed ? 'Expand' : 'Collapse'}
          >
            <Icon
              variant={isCollapsed ? 'CaretRightIcon' : 'CaretDownIcon'}
              size={12}
            />
          </button>
        ) : (
          <span />
        )}
        <SpanStatusDot status={status} />
        <span className="flex items-center gap-2 min-w-0">
          <Text family="mono" variant="subtext" nowrap as="span" className="truncate">
            {span.name}
          </Text>
          {span.attributes?.['nuon.tool'] ? (
            <Text variant="subtext" theme="neutral" nowrap as="span">
              · {span.attributes['nuon.tool']}
            </Text>
          ) : null}
        </span>
        <Text variant="subtext" family="mono" theme="neutral" nowrap as="span">
          {formatDurationNs(span.duration_ns)}
        </Text>
      </div>
      {hasChildren && !isCollapsed
        ? node.children.map((child) => (
            <SpanTreeNode
              key={child.span.span_id}
              node={child}
              collapsed={collapsed}
              onToggle={onToggle}
              selectedSpanId={selectedSpanId}
              onSelectSpan={onSelectSpan}
            />
          ))
        : null}
    </>
  )
}

type TSpanStatus = 'ok' | 'error' | 'skipped'

const statusFor = (node: TSpanNode): TSpanStatus => {
  const code = (node.span.status_code ?? '').toLowerCase()
  if (code === 'error' || node.hasErrorDescendant) return 'error'
  if (!node.span.end_time || node.span.duration_ns === 0) return 'skipped'
  return 'ok'
}

const SpanStatusDot = ({ status }: { status: TSpanStatus }) => {
  if (status === 'error') {
    return (
      <span
        title="Error (or contains errored descendant)"
        className="inline-flex items-center justify-center w-3 h-3 rounded-full bg-red-500 text-white text-[8px] leading-none"
      >
        ✗
      </span>
    )
  }
  if (status === 'skipped') {
    return (
      <span
        title="Skipped"
        className="inline-flex items-center justify-center w-3 h-3 rounded-full border border-cool-grey-400"
      />
    )
  }
  return (
    <span
      title="Ok"
      className="inline-flex items-center justify-center w-3 h-3 rounded-full bg-emerald-500"
    />
  )
}
