import { api } from '@/lib/api'
import type { TLogStream, TLogStreamLogsResponse } from '@/types/admin.types'

export const getLogStreams = (params: { search?: string }) =>
  api<{ log_streams: TLogStream[] }>({ path: 'log-streams', params })

export const getLogStreamDetail = (id: string) =>
  api<{ log_stream: TLogStream; logs: TLogStreamLogsResponse['logs']; page: number; total_pages: number }>({ path: `log-streams/${id}` })

export const getLogStreamLogs = (id: string, params: { page?: number }) =>
  api<TLogStreamLogsResponse>({ path: `log-streams/${id}/logs`, params })
