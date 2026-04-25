import { api } from '@/lib/api'
import type { TNamespaceWorkerInfo } from '@/types/admin.types'

export const getTemporalWorkers = () =>
  api<{ namespace_pollers: TNamespaceWorkerInfo[]; temporal_ui_url: string }>({ path: 'temporal-workers' })

export const getTemporalWorkerDetail = (namespace: string) =>
  api<{ info: TNamespaceWorkerInfo; temporal_ui_url: string }>({ path: `temporal-workers/${encodeURIComponent(namespace)}` })

export const getTemporalWorkflows = (params: { workflow_id?: string; run_id?: string; namespace?: string }) =>
  api<{ workflow_info: any }>({ path: 'temporal-workflows', params })
