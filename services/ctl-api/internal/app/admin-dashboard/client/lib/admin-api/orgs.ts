import { api } from '@/lib/api'
import type { TOrg, TOrgDetailResponse } from '@/types/admin.types'

export const getOrgs = (params: { search?: string; label?: string; page?: number }) =>
  api<{
    orgs: TOrg[]
    label_options: { key: string; values: string[] }[]
    page: number
    total_pages: number
  }>({ path: 'orgs', params })

export const getOrgDetail = (id: string, params?: { page?: number }) =>
  api<TOrgDetailResponse>({ path: `orgs/${id}`, params })

export const addOrgLabels = (id: string, labels: Record<string, string>) =>
  api<TOrg>({ path: `orgs/${id}/labels`, method: 'POST', body: { labels } })

export const removeOrgLabel = (id: string, key: string) =>
  api<TOrg>({ path: `orgs/${id}/labels/remove/${encodeURIComponent(key)}`, method: 'POST' })

export const addSupportUsers = (id: string) =>
  api<{ status: string }>({ path: `orgs/${id}/support-users/add`, method: 'POST' })

export const migrateOrgQueues = (id: string) =>
  api<{ status: string }>({ path: `orgs/${id}/migrate-queues`, method: 'POST' })

export const clearOrgQueues = (id: string) =>
  api<{ status: string; queues_cleared: number }>({ path: `orgs/${id}/clear-queues`, method: 'POST' })

export const forceRestartOrgQueues = (id: string) =>
  api<{ status: string; queues_restarted: number }>({ path: `orgs/${id}/force-restart-queues`, method: 'POST' })

export const removeOldRunnerProcesses = (id: string) =>
  api<{ status: string; processes_deleted: number }>({ path: `orgs/${id}/remove-old-runner-processes`, method: 'POST' })

export const shutdownOrgRunnerProcesses = (id: string) =>
  api<{ status: string; processes_shutdown: number; create_errors?: string[] }>({ path: `orgs/${id}/shutdown-runner-processes`, method: 'POST' })

export const shutdownHintOrgRunnerProcesses = (id: string) =>
  api<{ status: string; processes_shutdown: number }>({ path: `orgs/${id}/shutdown-hint-runner-processes`, method: 'POST' })
