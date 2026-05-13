import { useMemo, type ReactNode } from 'react'
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
  collapsed: Set<string>
  onToggleCollapsed: (spanId: string) => void
  onExpandAll: () => void
  onCollapseAll: () => void
  headerActions?: ReactNode
}

export const SpanTree = ({
  spans,
  selectedSpanId,
  onSelectSpan,
  collapsed,
  onToggleCollapsed,
  onExpandAll,
  onCollapseAll,
  headerActions,
}: ISpanTree) => {
  const forest = useMemo(() => buildSpanForest(spans), [spans])
  const isAllExpanded = collapsed.size === 0

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
    <div className="@container flex flex-col">
      <div className="flex items-center justify-between gap-3 px-2 h-14 border-b border-cool-grey-200 dark:border-dark-grey-600">
        <Text variant="body" theme="neutral">
          {spans.length} span{spans.length === 1 ? '' : 's'}
        </Text>
        <div className="flex items-center gap-2">
          {headerActions}
          <Button
            size="sm"
            variant="ghost"
            onClick={isAllExpanded ? onCollapseAll : onExpandAll}
            aria-pressed={!isAllExpanded}
            aria-label={isAllExpanded ? 'Collapse all' : 'Expand all'}
          >
            <Icon
              variant={isAllExpanded ? 'ArrowsInLineVerticalIcon' : 'ArrowsOutLineVerticalIcon'}
              size="12"
            />
            <span className="@max-[24rem]:hidden">{isAllExpanded ? 'Collapse' : 'Expand'}</span>
          </Button>
        </div>
      </div>
      <div className="flex flex-col">
        {forest.map((node) => (
          <SpanTreeNode
            key={node.span.span_id}
            node={node}
            collapsed={collapsed}
            onToggle={onToggleCollapsed}
            selectedSpanId={selectedSpanId}
            onSelectSpan={onSelectSpan}
          />
        ))}
      </div>
    </div>
  )
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
          'group flex items-stretch min-h-7 cursor-pointer',
          'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-500',
          isSelected &&
            'bg-primary-200 text-primary-800 dark:bg-primary-600/25 dark:text-primary-400'
        )}
        onClick={() => onSelectSpan(span.span_id)}
        title={span.name}
      >
        {Array.from({ length: node.depth }).map((_, i) => (
          <span
            key={i}
            aria-hidden="true"
            className="w-4 shrink-0 flex justify-center"
          >
            <span className="w-px self-stretch bg-cool-grey-200 dark:bg-dark-grey-700" />
          </span>
        ))}
        <div className="flex items-center gap-2 flex-1 min-w-0 px-2 py-1">
          {hasChildren ? (
            <button
              type="button"
              className={cn(
                'flex items-center justify-center w-4 h-4 shrink-0 rounded',
                'text-cool-grey-500 dark:text-cool-grey-400',
                'hover:bg-cool-grey-500/8 dark:hover:bg-cool-grey-500/8',
                'focus:outline-none focus-visible:ring-1 focus-visible:ring-primary-400'
              )}
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
            <span className="w-4 shrink-0" />
          )}
          <SpanStatusDot status={status} />
          <span className="flex items-center gap-2 min-w-0 flex-1">
            <Text
              family="mono"
              variant="subtext"
              nowrap
              as="span"
              className="truncate"
            >
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

const STATUS_ICON: Record<
  TSpanStatus,
  {
    variant: 'CheckCircleIcon' | 'XCircleIcon' | 'MinusCircleIcon'
    theme: 'success' | 'error' | 'neutral'
  }
> = {
  ok: { variant: 'CheckCircleIcon', theme: 'success' },
  error: { variant: 'XCircleIcon', theme: 'error' },
  skipped: { variant: 'MinusCircleIcon', theme: 'neutral' },
}

const STATUS_DOT_LABEL: Record<TSpanStatus, string> = {
  ok: 'Ok',
  error: 'Error (or contains errored descendant)',
  skipped: 'Skipped',
}

const SpanStatusDot = ({ status }: { status: TSpanStatus }) => {
  const { variant, theme } = STATUS_ICON[status]
  return (
    <span
      title={STATUS_DOT_LABEL[status]}
      aria-label={STATUS_DOT_LABEL[status]}
      className="inline-flex shrink-0"
    >
      <Icon variant={variant} weight="fill" size={14} theme={theme} />
    </span>
  )
}

export const collectSpanIds = (spans: TSpan[]): string[] => {
  const forest = buildSpanForest(spans)
  const out: string[] = []
  const walk = (n: TSpanNode) => {
    out.push(n.span.span_id)
    for (const c of n.children) walk(c)
  }
  for (const r of forest) walk(r)
  return out
}
