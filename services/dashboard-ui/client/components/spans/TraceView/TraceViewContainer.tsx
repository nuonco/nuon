import { useCallback, useEffect, useMemo, useRef } from 'react'
import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { useOrg } from '@/hooks/use-org'
import { useLogViewer } from '@/providers/log-viewer-provider'
import { getLogStreamSpans } from '@/lib'
import type { TSpan } from '@/types'
import { TraceView } from './TraceView'

export interface ITraceViewContainer {
  logStreamId: string
  shouldPoll?: boolean
  // When the parent already fetches spans (TracePanel does, so it can pass
  // the same list down to LogViewerProvider for parent-aggregation in the
  // span→logs cross-link), it forwards them here. If omitted, this container
  // falls back to fetching them itself — preserves the original standalone
  // usage.
  spans?: TSpan[]
  spansLoading?: boolean
}

// Container — orchestrates spans fetch + URL-driven span selection.
// Selecting a span pushes ?span_id=... onto the URL; the log viewer's
// client-side filters pick up that param and filter to that span.
// Cross-link in the other direction (log → span) works the same way:
// landing on the trace tab with ?span_id=... preselects the matching span.
export const TraceViewContainer = ({
  logStreamId,
  shouldPoll,
  spans: propSpans,
  spansLoading,
}: ITraceViewContainer) => {
  const { org } = useOrg()
  const [searchParams, setSearchParams] = useSearchParams()
  const selectedSpanId = searchParams.get('span_id') ?? undefined

  const { data: fetchedSpans, isLoading: fetchedLoading } = useQuery({
    queryKey: ['log-stream-spans', org?.id, logStreamId],
    queryFn: () => getLogStreamSpans({ orgId: org.id, logStreamId }),
    enabled: !!org?.id && !!logStreamId && propSpans === undefined,
    refetchInterval: shouldPoll ? 5000 : false,
  })

  const spans = propSpans ?? fetchedSpans
  const isLoading = propSpans !== undefined ? !!spansLoading : fetchedLoading

  const sortedSpans = useMemo<TSpan[]>(() => spans ?? [], [spans])

  const handleSelectSpan = useCallback(
    (spanId: string | undefined) => {
      const next = new URLSearchParams(searchParams)
      if (spanId) next.set('span_id', spanId)
      else next.delete('span_id')
      setSearchParams(next, { replace: true })
    },
    [searchParams, setSearchParams]
  )

  useEffect(() => {
    return () => {
      const params = new URLSearchParams(window.location.search)
      if (params.has('span_id')) {
        params.delete('span_id')
        setSearchParams(params, { replace: true })
      }
    }
  }, [])

  // Cross-link: when the user clicks a log row in the right pane, the
  // LogViewerProvider exposes the active log via useLogViewer. Each log
  // record already carries span_id from otelzap, so we mirror it onto
  // the URL — this both highlights the matching node in the tree and
  // narrows the logs pane to that span.
  const { activeLog } = useLogViewer()
  const lastSyncedSpanRef = useRef<string | undefined>()
  useEffect(() => {
    const next = activeLog?.span_id
    if (!next || next === lastSyncedSpanRef.current) return
    lastSyncedSpanRef.current = next
    handleSelectSpan(next)
  }, [activeLog, handleSelectSpan])

  return (
    <TraceView
      spans={sortedSpans}
      isLoading={isLoading}
      selectedSpanId={selectedSpanId}
      onSelectSpan={handleSelectSpan}
      rightPane={<SSELogs filterClassName="top-0" />}
    />
  )
}
