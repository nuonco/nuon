import { api } from '@/lib/api'
import type { TMe } from '@/types'

export const getMe = () =>
  api<TMe>({
    path: `auth/me`,
  })