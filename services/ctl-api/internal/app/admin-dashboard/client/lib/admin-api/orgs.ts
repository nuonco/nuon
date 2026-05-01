import { api } from '@/lib/api'
import type { TOrg, TOrgsResponse, TOrgDetailResponse } from '@/types/admin.types'

export const getOrgs = (params: { search?: string; tag?: string[]; page?: number }) =>
  api<TOrgsResponse>({ path: 'orgs', params })

export const getOrgDetail = (id: string, params?: { page?: number }) =>
  api<TOrgDetailResponse>({ path: `orgs/${id}`, params })

export const updateOrgTags = (id: string, tags: string[]) =>
  api<TOrg>({ path: `orgs/${id}/tags`, method: 'POST', body: { tags } })

export const removeOrgTag = (id: string, tag: string) =>
  api<TOrg>({ path: `orgs/${id}/tags/remove/${encodeURIComponent(tag)}`, method: 'POST' })

export const addOrgLabels = (id: string, labels: Record<string, string>) =>
  api<TOrg>({ path: `orgs/${id}/labels`, method: 'POST', body: { labels } })

export const removeOrgLabel = (id: string, key: string) =>
  api<TOrg>({ path: `orgs/${id}/labels/remove/${encodeURIComponent(key)}`, method: 'POST' })

export const addSupportUsers = (id: string) =>
  api<{ status: string }>({ path: `orgs/${id}/support-users/add`, method: 'POST' })

export const migrateOrgQueues = (id: string) =>
  api<{ status: string }>({ path: `orgs/${id}/migrate-queues`, method: 'POST' })
