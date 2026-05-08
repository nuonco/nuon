import { api } from '@/lib/api'

export type TQueryRecord = {
  sql: string
  table: string
  operation: string
  db_type: string
  source: string
  endpoint: string
  count: number
  total_ms: number
  avg_ms: number
  min_ms: number
  max_ms: number
  total_rows: number
  max_rows: number
  max_response_size: number
  last_error?: string
  caller: string
  caller_url?: string
  last_seen_at: string
}

export const getQueries = (params?: {
  search?: string
  table?: string
  db_type?: string
  source?: string
  has_error?: string
  sort?: string
  min_duration_ms?: string
  time_range?: string
}) =>
  api<{ enabled: boolean; queries: TQueryRecord[]; tables: string[]; total: number }>({
    path: 'queries',
    params,
  })

export const explainQuery = (params: { sql: string; db_type: string }) =>
  api<{ rows: Record<string, unknown>[]; error?: string }>({
    path: 'queries/explain',
    method: 'POST',
    body: params,
  })
