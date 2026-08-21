import { api } from '@/lib/api'
import type { TFeatureFlag } from '@/types/admin.types'

export const getFeatureFlags = () =>
  api<{ flags: TFeatureFlag[]; total_orgs: number }>({ path: 'feature-flags' })
