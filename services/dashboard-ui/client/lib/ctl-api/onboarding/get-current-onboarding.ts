import { api } from '@/lib/api'
import type { TOnboarding } from '@/types'

export const getCurrentOnboarding = () =>
  api<TOnboarding>({
    path: `onboarding/current`,
  })
