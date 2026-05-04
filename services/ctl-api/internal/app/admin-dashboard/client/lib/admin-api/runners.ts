import { api } from '@/lib/api'
import type { TRunnersResponse, TRunnerDetailView, TAllRunnersResponse } from '@/types/admin.types'

export const getRunners = () =>
  api<TRunnersResponse>({ path: 'runners' })

export const getAllRunners = (params?: { org_id?: string; page?: number }) =>
  api<TAllRunnersResponse>({ path: 'runners/all', params })

export const getRunnerDetail = (id: string) =>
  api<TRunnerDetailView>({ path: `runners/${id}` })

export const upsertRunnerConfig = (id: string, body: { job_type: string; duration: number; should_error: boolean; panic: boolean; trigger_shutdown: boolean }) =>
  api<{ status: string }>({ path: `runners/${id}/configs`, method: 'PUT', body })

export const deleteRunnerConfig = (id: string, jobType: string) =>
  api<{ status: string }>({ path: `runners/${id}/configs/${encodeURIComponent(jobType)}`, method: 'DELETE' })

export const resetRunnerConfigs = (id: string) =>
  api<{ status: string }>({ path: `runners/${id}/configs/reset`, method: 'POST' })
