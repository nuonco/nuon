import { api } from '@/lib/api'
import type { TOTELLog } from '@/types'

export type TLogStreamFilters = {
  service_name?: string[]
  scope_name?: string[]
  severity_text?: string[]
  tool?: string
  helm_release_name?: string
  helm_operation?: string
  tf_workspace_id?: string
  tf_operation?: string
  k8s_kind?: string
  k8s_namespace?: string
  k8s_name?: string
  q?: string
  // Phase 2 — span/trace filtering. The runner emits otelzap log records
  // whose trace context populates these columns directly, so the API can
  // narrow log results to a single span / trace.
  span_id?: string
  trace_id?: string
}

const buildLogQueryString = ({
  order,
  filters,
}: {
  order: 'asc' | 'desc'
  filters?: TLogStreamFilters
}): string => {
  const sp = new URLSearchParams()
  sp.set('order', order)
  if (filters) {
    for (const [key, raw] of Object.entries(filters)) {
      if (raw == null) continue
      if (Array.isArray(raw)) {
        for (const v of raw) {
          if (v != null && v !== '') sp.append(key, String(v))
        }
      } else if (raw !== '') {
        sp.append(key, String(raw))
      }
    }
  }
  const s = sp.toString()
  return s ? `?${s}` : ''
}

export const getLogStreamLogs = ({
  logStreamId,
  orgId,
  offset,
  order = 'asc',
  filters,
}: {
  logStreamId: string
  orgId: string
  offset?: string
  order?: 'asc' | 'desc'
  filters?: TLogStreamFilters
}) =>
  api<TOTELLog[]>({
    path: `log-streams/${logStreamId}/logs${buildLogQueryString({ order, filters })}`,
    orgId,
    headers: offset ? { 'X-Nuon-API-Offset': offset } : {},
  })

export const getLogStreamLogsWithMeta = ({
  logStreamId,
  orgId,
  offset,
  order = 'asc',
  filters,
}: {
  logStreamId: string
  orgId: string
  offset?: string
  order?: 'asc' | 'desc'
  filters?: TLogStreamFilters
}) =>
  api<TOTELLog[]>({
    path: `log-streams/${logStreamId}/logs${buildLogQueryString({ order, filters })}`,
    orgId,
    headers: offset ? { 'X-Nuon-API-Offset': offset } : {},
    withMeta: true,
  })
