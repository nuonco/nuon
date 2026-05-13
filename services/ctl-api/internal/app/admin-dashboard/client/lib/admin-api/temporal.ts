import { api } from '@/lib/api'
import type { TNamespaceWorkerInfo } from '@/types/admin.types'

export const getTemporalWorkers = () =>
  api<{ namespace_pollers: TNamespaceWorkerInfo[]; temporal_ui_url: string }>({ path: 'temporal-workers' })

export const getTemporalWorkerDetail = (namespace: string) =>
  api<{ info: TNamespaceWorkerInfo; temporal_ui_url: string }>({ path: `temporal-workers/${encodeURIComponent(namespace)}` })

export const getTemporalWorkflows = (params: { workflow_id?: string; run_id?: string; namespace?: string }) =>
  api<{ workflow_info: any; temporal_ui_url: string; namespace: string; workflow_id: string; run_id: string }>({ path: 'temporal-workflows', params })

export const getTemporalWorkflowNamespaces = () =>
  api<{ namespaces: string[]; temporal_ui_url: string }>({ path: 'temporal-workflows/namespaces' })

export const getTemporalWorkflowStats = (params: { namespace: string; workflow_id: string }) =>
  api<{ history_length: number; history_size_bytes: number; can_count: number; start_time: string; status: string }>({ path: 'temporal-workflows/stats', params })
