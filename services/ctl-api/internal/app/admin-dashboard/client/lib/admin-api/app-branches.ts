import { api } from '@/lib/api'
import type {
  TAppBranchesResponse,
  TAppBranchDetailResponse,
  TAppBranchRunsResponse,
  TAppBranchWorkflowsResponse,
} from '@/types/admin.types'

export const getAppBranches = (params: {
  search?: string
  sort?: string
  managed_by?: string
  show_deleted?: string
  page?: number
}) => api<TAppBranchesResponse>({ path: 'app-branches', params })

export const getAppBranchDetail = (id: string) =>
  api<TAppBranchDetailResponse>({ path: `app-branches/${id}` })

export const getAppBranchRuns = (id: string, params: { page?: number }) =>
  api<TAppBranchRunsResponse>({ path: `app-branches/${id}/runs`, params })

export const getAppBranchWorkflows = (id: string, params: { page?: number }) =>
  api<TAppBranchWorkflowsResponse>({ path: `app-branches/${id}/workflows`, params })
