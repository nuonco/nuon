import { api } from '@/lib/api'

export type TWorkflowFilters = {
  search?: string
  type?: string
  status?: string
  created_after?: string
  created_before?: string
}

export type TBulkCancelResponse = {
  status: 'enqueued' | 'noop'
  count: number
  message: string
  queue_id?: string
  signal_id?: string
}

export type TWorkflowFilterOptions = {
  types: string[]
  cancelable_statuses: { value: string; label: string }[]
  per_page_options: number[]
}

export const getWorkflows = (params: TWorkflowFilters & { sort?: string; page?: number; per_page?: number }) =>
  api<{ workflows: any[]; page: number; per_page: number; total: number; total_pages: number }>({ path: 'workflows', params })

export const getWorkflowFilterOptions = () =>
  api<TWorkflowFilterOptions>({ path: 'workflows/filter-options' })

export type TWorkflowTypeStat = {
  type: string
  count: number
}

export const getWorkflowTypeStats = (params: Omit<TWorkflowFilters, 'type'>) =>
  api<{ stats: TWorkflowTypeStat[] }>({ path: 'workflows/type-stats', params })

export const getWorkflowDetail = (workflowId: string) =>
  api<{ workflow: any; group_details: any[]; generate_steps_signal: any; workflow_signal: any; workflow_info?: any }>({ path: `workflows/${workflowId}` })

export const bulkCancelWorkflows = (body: TWorkflowFilters & { workflow_ids?: string[] }) =>
  api<TBulkCancelResponse>({ path: 'workflows/bulk-cancel', method: 'POST', body })
