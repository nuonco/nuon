import { api } from '@/lib/api'
import type { TTemporalWorkersResponse, TNamespaceWorkerInfo } from '@/types/admin.types'

export const getTemporalWorkers = () =>
  api<TTemporalWorkersResponse>({ path: 'temporal-workers' })

export const getTemporalWorkerDetail = (namespace: string) =>
  api<TNamespaceWorkerInfo>({ path: `temporal-workers/${encodeURIComponent(namespace)}` })

export const getTemporalWorkflows = (params: { workflow_id?: string; run_id?: string; namespace?: string }) =>
  api<{ workflow_info: any }>({ path: 'temporal-workflows', params })
