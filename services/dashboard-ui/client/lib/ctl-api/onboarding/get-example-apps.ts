import { api } from '@/lib/api'
import type { TExampleApp } from '@/types'

export const getExampleApps = () =>
  api<TExampleApp[]>({
    path: `onboarding/example-apps`,
  })
