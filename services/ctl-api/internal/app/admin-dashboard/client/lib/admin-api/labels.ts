import { api } from '@/lib/api'
import type { TLabelsResponse } from '@/types/admin.types'

export const getLabels = () =>
  api<TLabelsResponse>({ path: 'labels' })
