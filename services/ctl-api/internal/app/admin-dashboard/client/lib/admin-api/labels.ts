import { api } from '@/lib/api'
import type { TLabelsResponse } from '@/types/admin.types'

export const getLabels = (params?: { search?: string; entity_type?: string; org_id?: string; page?: number }) =>
  api<TLabelsResponse>({ path: 'labels', params })
