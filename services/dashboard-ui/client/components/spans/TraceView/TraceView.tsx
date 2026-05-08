import { useMemo, useState, type ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { SpanTree, collectSpanIds } from '@/components/spans/SpanTree/SpanTree'
import { TimelineBars } from '@/components/spans/TimelineBars'
import type { TSpan } from '@/types'
import { cn } from '@/utils/classnames'

export type TTraceRightPaneVariant = 'logs' | 'timeline'

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
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())

  const allIds = useMemo(() => collectSpanIds(spans), [spans])

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

  const toggle = (
    <ToggleButton<TTraceRightPaneVariant>
      value={variant}
      onChange={setVariant}
      options={[
        {
          value: 'logs',
          label: (
            <>
              <Icon variant="ListIcon" size="12" />
              <span className="@max-[24rem]:hidden">Logs</span>
            </>
          ),
          ariaLabel: 'Show logs',
        },
        {
          value: 'timeline',
          label: (
            <>
              <Icon variant="TimerIcon" size="12" />
              <span className="@max-[24rem]:hidden">Timeline</span>
            </>
          ),
          ariaLabel: 'Show timeline',
        },
      ]}
    />
  )

  return (
    <div className="flex flex-col gap-4 h-full">
      {showUpgradeBanner ? (
        <Banner theme="info">
          <Text weight="strong">No trace data available</Text>
          <Text variant="subtext">
            Traces require a recent version of the runner. If this run completed
            without spans, upgrade your runner to see execution traces here.
          </Text>
        </Banner>
      ) : null}
      <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,3fr)_minmax(0,7fr)] flex-auto min-h-0">
        <div className="overflow-y-auto min-h-[20rem]">
          {isLoading && !spans.length ? (
            <div className="p-6 text-center">
              <Text variant="subtext" theme="neutral">
                Loading spans…
              </Text>
            </div>
          ) : (
            <SpanTree
              spans={spans}
              selectedSpanId={selectedSpanId}
              onSelectSpan={handleSelect}
              collapsed={collapsed}
              onToggleCollapsed={handleToggleCollapsed}
              onExpandAll={handleExpandAll}
              onCollapseAll={handleCollapseAll}
              headerActions={toggle}
            />
          )}
        </div>
        <div
          className={cn(
            'overflow-y-auto min-h-[20rem] lg:border-l lg:border-cool-grey-200 lg:dark:border-dark-grey-600',
            isLogs && 'px-3'
          )}
        >
          {isLogs ? (
            rightPane
          ) : (
            <TimelineBars
              spans={spans}
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
