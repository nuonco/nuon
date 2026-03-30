import { api } from '@/lib/api'
import type { TOnboarding } from '@/types'

export const createOnboarding = () =>
  api<TOnboarding>({
    method: 'POST',
    path: `onboarding`,
  })
