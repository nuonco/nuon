import { api } from '@/lib/api'
import type { TSpan } from '@/types'

// Wire shape from ctl-api (services/ctl-api/internal/app/runners/service/
// log_stream_read_spans.go). Field names differ from TSpan — we remap below
// so consumers don't have to think about the protocol surface.
type TLogStreamSpanWire = {
  span_id: string
  parent_span_id: string
  trace_id: string
  span_name: string
  span_kind?: string
  service_name?: string
  scope_name?: string
  status_code?: string
  status_message?: string
  start_timestamp: string
  end_timestamp: string
  duration_ns: number
  attributes?: Record<string, string>
}

const fromWire = (w: TLogStreamSpanWire): TSpan => ({
  span_id: w.span_id,
  parent_span_id: w.parent_span_id || undefined,
  trace_id: w.trace_id,
  name: w.span_name,
  start_time: w.start_timestamp,
  end_time: w.end_timestamp,
  duration_ns: w.duration_ns,
  status_code: w.status_code,
  status_message: w.status_message,
  attributes: w.attributes,
  scope_name: w.scope_name,
  service_name: w.service_name,
})

// Phase 4 endpoint: GET /v1/log-streams/:id/spans
// Returns a flat span list. The frontend assembles the tree from
// parent_span_id (siblings sorted by start_time, root has empty parent).
//
// While Phase 4 is in flight the endpoint may not exist yet. We swallow
// 404s into an empty array so the trace tab can render its empty state
// instead of throwing — once the endpoint ships this becomes a normal
// passthrough.
export const getLogStreamSpans = async ({
  logStreamId,
  orgId,
}: {
  logStreamId: string
  orgId: string
}): Promise<TSpan[]> => {
  try {
    const wire = await api<TLogStreamSpanWire[]>({
      path: `log-streams/${logStreamId}/spans`,
      orgId,
    })
    return (wire ?? []).map(fromWire)
  } catch (err: any) {
    if (err?.status === 404 || err?.statusCode === 404) {
      return []
    }
    throw err
  }
}
