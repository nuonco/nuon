import { api } from '@/lib/api'
import type { TOnboarding } from '@/types'

export const abandonOnboarding = () =>
  api<TOnboarding>({
    method: 'DELETE',
    path: `onboarding/current`,
  })
