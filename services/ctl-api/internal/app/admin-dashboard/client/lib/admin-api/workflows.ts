import { api } from '@/lib/api'
import type { TWorkflowsResponse, TWorkflowDetailResponse } from '@/types/admin.types'

export const getWorkflows = (params: { search?: string; page?: number }) =>
  api<TWorkflowsResponse>({ path: 'workflows', params })

export const getWorkflowDetail = (workflowId: string) =>
  api<TWorkflowDetailResponse>({ path: `workflows/${workflowId}` })
