import { api } from '@/lib/api'
import type {
  TInstallsResponse,
  TInstallDetailResponse,
  TInstallActiveDeploymentsResponse,
  TInstallActivityResponse,
  TInstallWorkflowsResponse,
  TInstall,
} from '@/types/admin.types'

export const getInstalls = (params: { search?: string; creator_type?: string; sort?: string; show_deleted?: string; page?: number }) => {
  const { creator_type, show_deleted, ...rest } = params
  return api<TInstallsResponse>({
    path: 'installs',
    params: { ...rest, filter: creator_type, deleted_filter: show_deleted },
  })
}

export const getInstallDetail = (id: string) =>
  api<TInstallDetailResponse>({ path: `installs/${id}` })

export const getInstallRunnerStatus = (id: string) =>
  api<{ status: string; description: string }>({ path: `installs/${id}/status/runner` })

export const getInstallSandboxStatus = (id: string) =>
  api<{ status: string; description: string }>({ path: `installs/${id}/status/sandbox` })

export const getInstallComponentStatus = (id: string) =>
  api<{ status: string; description: string }>({ path: `installs/${id}/status/component` })

export const getInstallDriftStatus = (id: string) =>
  api<{ status: string; description: string }>({ path: `installs/${id}/status/drift` })

export const getInstallActiveDeployments = (id: string) =>
  api<TInstallActiveDeploymentsResponse>({ path: `installs/${id}/active-deployments` })

export const getInstallActivity = (id: string, params: { page?: number; entity_type?: string; start_date?: string; end_date?: string }) => {
  const { entity_type, ...rest } = params
  return api<TInstallActivityResponse>({
    path: `installs/${id}/activity`,
    params: { ...rest, entity_types: entity_type },
  })
}

export const getInstallWorkflows = (id: string, params: { page?: number }) =>
  api<TInstallWorkflowsResponse>({ path: `installs/${id}/workflows`, params })

export const addInstallLabel = (id: string, key: string, value: string) =>
  api<TInstall>({ path: `installs/${id}/labels`, method: 'POST', body: { key, value } })

export const removeInstallLabel = (id: string, key: string) =>
  api<TInstall>({ path: `installs/${id}/labels/remove/${encodeURIComponent(key)}`, method: 'POST' })
