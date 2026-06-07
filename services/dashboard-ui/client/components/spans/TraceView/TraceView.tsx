import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { SpanTree, collectSpanIds } from '@/components/spans/SpanTree/SpanTree'
import { TimelineBars } from '@/components/spans/TimelineBars'
import type { TSpan } from '@/types'
import { cn } from '@/utils/classnames'
import { filterRunnerInternal } from '@/utils/span-tree'

export type TTraceRightPaneVariant = 'logs' | 'timeline'
export type TTraceScopeVariant = 'user' | 'all'

export interface ITraceView {
  spans: TSpan[]
  isLoading?: boolean
  selectedSpanId?: string
  onSelectSpan: (spanId: string | undefined) => void
  rightPane: ReactNode
}

export const TraceView = ({
  spans,
  isLoading,
  selectedSpanId,
  onSelectSpan,
  rightPane,
}: ITraceView) => {
  const [variant, setVariant] = useState<TTraceRightPaneVariant>('logs')
  const [scope, setScope] = useState<TTraceScopeVariant>('user')
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())

  // Resizable left/right split (lg+ only). Width of the left (span tree) pane,
  // expressed as a percentage of the container and clamped to a usable range.
  const splitRef = useRef<HTMLDivElement>(null)
  const [leftWidthPct, setLeftWidthPct] = useState(30)

  const updateWidthFromClientX = useCallback((clientX: number) => {
    const el = splitRef.current
    if (!el) return
    const rect = el.getBoundingClientRect()
    if (rect.width === 0) return
    const pct = ((clientX - rect.left) / rect.width) * 100
    setLeftWidthPct(Math.min(80, Math.max(20, pct)))
  }, [])

  const handlePointerMove = useCallback(
    (e: PointerEvent) => updateWidthFromClientX(e.clientX),
    [updateWidthFromClientX]
  )

  const stopResizing = useCallback(() => {
    window.removeEventListener('pointermove', handlePointerMove)
    window.removeEventListener('pointerup', stopResizing)
    document.body.style.userSelect = ''
    document.body.style.cursor = ''
  }, [handlePointerMove])

  const startResizing = useCallback(
    (e: React.PointerEvent) => {
      e.preventDefault()
      window.addEventListener('pointermove', handlePointerMove)
      window.addEventListener('pointerup', stopResizing)
      document.body.style.userSelect = 'none'
      document.body.style.cursor = 'col-resize'
    },
    [handlePointerMove, stopResizing]
  )

  const handleSeparatorKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'ArrowLeft') {
      e.preventDefault()
      setLeftWidthPct((p) => Math.max(20, p - 2))
    } else if (e.key === 'ArrowRight') {
      e.preventDefault()
      setLeftWidthPct((p) => Math.min(80, p + 2))
    }
  }, [])

  // Clean up window listeners if we unmount mid-drag.
  useEffect(() => () => stopResizing(), [stopResizing])

  const visibleSpans = useMemo(
    () => (scope === 'user' ? filterRunnerInternal(spans) : spans),
    [spans, scope]
  )
  const hiddenCount = spans.length - visibleSpans.length

  const allIds = useMemo(() => collectSpanIds(visibleSpans), [visibleSpans])

  const handleToggleCollapsed = (id: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const handleExpandAll = () => setCollapsed(new Set())
  const handleCollapseAll = () => setCollapsed(new Set(allIds))

  const handleSelect = (id: string) =>
    onSelectSpan(id === selectedSpanId ? undefined : id)

  const showUpgradeBanner = !isLoading && spans.length === 0
  const isLogs = variant === 'logs'

  const headerActions = (
    <>
      <ToggleButton<TTraceScopeVariant>
        value={scope}
        onChange={setScope}
        options={[
          {
            value: 'user',
            label: (
              <>
                <Icon variant="UserIcon" size="12" />
                <span className="@max-[30rem]:hidden">User</span>
              </>
            ),
            ariaLabel:
              hiddenCount > 0
                ? `Show user actions only (${hiddenCount} runner spans hidden)`
                : 'Show user actions only',
          },
          {
            value: 'all',
            label: (
              <>
                <Icon variant="StackIcon" size="12" />
                <span className="@max-[30rem]:hidden">All</span>
              </>
            ),
            ariaLabel: 'Show all spans including runner internals',
          },
        ]}
      />
      <ToggleButton<TTraceRightPaneVariant>
        value={variant}
        onChange={setVariant}
        options={[
          {
            value: 'logs',
            label: (
              <>
                <Icon variant="ListIcon" size="12" />
                <span className="@max-[30rem]:hidden">Logs</span>
              </>
            ),
            ariaLabel: 'Show logs',
          },
          {
            value: 'timeline',
            label: (
              <>
                <Icon variant="TimerIcon" size="12" />
                <span className="@max-[30rem]:hidden">Timeline</span>
              </>
            ),
            ariaLabel: 'Show timeline',
          },
        ]}
      />
    </>
  )

  return (
    <div className="flex flex-col gap-4 h-full">
      {showUpgradeBanner ? (
        <Banner theme="info" className="mt-3">
          <Text weight="strong">No trace data available</Text>
          <Text variant="subtext" className="!block">
            Traces require a recent version of the runner. If this run completed
            without spans, upgrade your runner to see execution traces here.
          </Text>
        </Banner>
      ) : null}
      <div
        ref={splitRef}
        className="flex flex-col lg:flex-row flex-auto min-h-0"
        style={{ '--trace-left': `${leftWidthPct}%` } as React.CSSProperties}
      >
        <div className="overflow-y-auto min-h-[20rem] w-full lg:w-[var(--trace-left)] lg:shrink-0">
          {isLoading && !spans.length ? (
            <div className="p-6 text-center">
              <Text variant="subtext" theme="neutral">
                Loading spans…
              </Text>
            </div>
          ) : (
            <SpanTree
              spans={visibleSpans}
              hiddenCount={scope === 'user' ? hiddenCount : 0}
              selectedSpanId={selectedSpanId}
              onSelectSpan={handleSelect}
              collapsed={collapsed}
              onToggleCollapsed={handleToggleCollapsed}
              onExpandAll={handleExpandAll}
              onCollapseAll={handleCollapseAll}
              headerActions={headerActions}
            />
          )}
        </div>

        <div
          role="separator"
          aria-orientation="vertical"
          aria-label="Resize panels"
          tabIndex={0}
          onPointerDown={startResizing}
          onKeyDown={handleSeparatorKeyDown}
          className="group hidden lg:flex shrink-0 w-1.5 cursor-col-resize items-stretch justify-center touch-none focus:outline-none"
        >
          <span className="w-px bg-cool-grey-200 dark:bg-dark-grey-600 transition-colors group-hover:bg-primary-400 group-focus-visible:bg-primary-500" />
        </div>

        <div
          className={cn(
            'overflow-y-auto min-h-[20rem] w-full lg:flex-1 lg:min-w-0',
            isLogs && 'px-3'
          )}
        >
          {isLogs ? (
            rightPane
          ) : (
            <TimelineBars
              spans={visibleSpans}
              selectedSpanId={selectedSpanId}
              onSelectSpan={handleSelect}
              collapsed={collapsed}
            />
          )}
        </div>
      </div>
    </div>
  )
}
