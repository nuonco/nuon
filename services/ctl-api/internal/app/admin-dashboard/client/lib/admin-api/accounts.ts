import { api } from '@/lib/api'
import type { TAccountsResponse, TAccountDetailResponse, TInstallsResponse, TInstallActivityResponse } from '@/types/admin.types'

export const getAccounts = (params: { search?: string; account_type?: string; page?: number }) => {
  const { account_type, ...rest } = params
  return api<TAccountsResponse>({ path: 'accounts', params: { ...rest, filter: account_type } })
}

export const getAccountDetail = (id: string) =>
  api<TAccountDetailResponse>({ path: `accounts/${id}` })

export const getAccountInstalls = (id: string, params: { page?: number }) =>
  api<TInstallsResponse>({ path: `accounts/${id}/installs`, params })

export const getAccountAuditLogs = (id: string, params: { page?: number; entity_type?: string; start_date?: string; end_date?: string }) => {
  const { entity_type, ...rest } = params
  return api<TInstallActivityResponse>({ path: `accounts/${id}/audit-logs`, params: { ...rest, entity_types: entity_type } })
}
