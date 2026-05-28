import { Fragment, useMemo, useState, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TSpan } from '@/types'
import { cn } from '@/utils/classnames'
import { buildSpanForest, formatDurationNs, type TSpanNode } from '@/utils/span-tree'

export interface ISpanTree {
  spans: TSpan[]
  // When the parent applied a span filter (e.g. the "User actions" toggle in
  // TraceView), this is the count of spans dropped from `spans`. Surfaced in
  // the header so users can tell something is being hidden.
  hiddenCount?: number
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
  hiddenCount = 0,
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
    <div className="flex flex-col">
      <div className="flex items-center justify-between gap-3 px-2 h-14 border-b border-cool-grey-200 dark:border-dark-grey-600">
        <Text variant="subtext" theme="neutral" className="whitespace-nowrap shrink-0">
          {spans.length} span{spans.length === 1 ? '' : 's'}
          {hiddenCount > 0 ? (
            <Text variant="subtext" theme="neutral" as="span">
              {' '}({hiddenCount} hidden)
            </Text>
          ) : null}
        </Text>
        <div className="@container flex flex-auto justify-end items-center gap-2">
          {headerActions}
          <Button
            size="sm"
            onClick={isAllExpanded ? onCollapseAll : onExpandAll}
            aria-pressed={!isAllExpanded}
            aria-label={isAllExpanded ? 'Collapse all' : 'Expand all'}
          >
            <Icon
              variant={isAllExpanded ? 'ArrowsInLineVerticalIcon' : 'ArrowsOutLineVerticalIcon'}
              size="12"
            />
            <span className="@max-[14rem]:hidden">{isAllExpanded ? 'Collapse' : 'Expand'}</span>
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

  // Per-row toggle for the inline attribute panel. Local state because the
  // panel is purely a UI affordance — no need to lift to the parent or
  // mirror onto the URL like span selection does.
  const [showAttrs, setShowAttrs] = useState(false)
  const attrEntries = useMemo(
    () => (span.attributes ? Object.entries(span.attributes).sort() : []),
    [span.attributes]
  )
  const hasAttrs = attrEntries.length > 0

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
          {hasAttrs ? (
            <button
              type="button"
              className={cn(
                'flex items-center justify-center w-4 h-4 shrink-0 rounded',
                'text-cool-grey-500 dark:text-cool-grey-400',
                'hover:bg-cool-grey-500/8 dark:hover:bg-cool-grey-500/8',
                'focus:outline-none focus-visible:ring-1 focus-visible:ring-primary-400',
                showAttrs && 'text-primary-600 dark:text-primary-400'
              )}
              onClick={(e) => {
                e.stopPropagation()
                setShowAttrs((v) => !v)
              }}
              aria-pressed={showAttrs}
              aria-label={showAttrs ? 'Hide attributes' : 'Show attributes'}
            >
              <Icon variant="InfoIcon" size={12} />
            </button>
          ) : null}
          <Text variant="subtext" family="mono" theme="neutral" nowrap as="span">
            {formatDurationNs(span.duration_ns)}
          </Text>
        </div>
      </div>
      {hasAttrs && showAttrs ? (
        <SpanAttributePanel depth={node.depth} entries={attrEntries} />
      ) : null}
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

interface ISpanAttributePanel {
  depth: number
  entries: [string, string][]
}

const SpanAttributePanel = ({ depth, entries }: ISpanAttributePanel) => (
  <div className="flex items-stretch">
    {Array.from({ length: depth }).map((_, i) => (
      <span
        key={i}
        aria-hidden="true"
        className="w-4 shrink-0 flex justify-center"
      >
        <span className="w-px self-stretch bg-cool-grey-200 dark:bg-dark-grey-700" />
      </span>
    ))}
    <div className="flex-1 min-w-0 px-3 py-2 ml-4 border-l border-cool-grey-200 dark:border-dark-grey-700 bg-cool-grey-50 dark:bg-dark-grey-700/30">
      <dl className="grid grid-cols-[max-content_minmax(0,1fr)] gap-x-3 gap-y-1">
        {entries.map(([k, v]) => (
          <Fragment key={k}>
            <dt>
              <Text family="mono" variant="subtext" theme="neutral" as="span" nowrap>
                {k}
              </Text>
            </dt>
            <dd className="min-w-0">
              <Text family="mono" variant="subtext" as="span" className="break-all">
                {v}
              </Text>
            </dd>
          </Fragment>
        ))}
      </dl>
    </div>
  </div>
)

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
