import { api } from '@/lib/api'

export const getWorkflows = (params: { search?: string; type?: string; sort?: string; page?: number }) =>
  api<{ workflows: any[]; page: number; total_pages: number }>({ path: 'workflows', params })

export const getWorkflowDetail = (workflowId: string) =>
  api<{ workflow: any; group_details: any[]; generate_steps_signal: any; workflow_signal: any; workflow_info?: any }>({ path: `workflows/${workflowId}` })
