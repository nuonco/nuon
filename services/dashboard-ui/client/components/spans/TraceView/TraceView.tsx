import { useState, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { SpanTimeline } from '@/components/spans/SpanTimeline'
import { SpanTree } from '@/components/spans/SpanTree'
import type { TSpan } from '@/types'

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

  return (
    <div className="flex flex-col gap-4 h-full">
      <div className="flex items-center justify-between gap-2 px-2">
        <div className="flex items-center gap-2">
          <Text variant="base" weight="strong">
            Trace
          </Text>
          {spans.length > 0 ? (
            <Text variant="subtext" theme="neutral">
              {spans.length} span{spans.length === 1 ? '' : 's'}
            </Text>
          ) : null}
        </div>
        <div className="flex items-center gap-1">
          <Button
            size="xs"
            variant={variant === 'logs' ? 'primary' : 'ghost'}
            onClick={() => setVariant('logs')}
          >
            Logs
          </Button>
          <Button
            size="xs"
            variant={variant === 'timeline' ? 'primary' : 'ghost'}
            onClick={() => setVariant('timeline')}
          >
            <Icon variant="ArrowsLeftRightIcon" size={12} /> Split
          </Button>
        </div>
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)] gap-4 flex-auto min-h-0">
        <div className="border rounded overflow-y-auto min-h-[20rem]">
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
              onSelectSpan={(id) =>
                onSelectSpan(id === selectedSpanId ? undefined : id)
              }
            />
          )}
        </div>
        <div className="border rounded overflow-y-auto min-h-[20rem]">
          {variant === 'logs' ? (
            rightPane
          ) : (
            <SpanTimeline
              spans={spans}
              selectedSpanId={selectedSpanId}
              onSelectSpan={(id) =>
                onSelectSpan(id === selectedSpanId ? undefined : id)
              }
            />
          )}
        </div>
      </div>
    </div>
  )
}
