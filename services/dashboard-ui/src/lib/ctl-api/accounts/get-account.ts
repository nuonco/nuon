import { api } from '@/lib/api'
import type { TAccount } from '@/types'

export const getAccount = () =>
  api<TAccount>({
    path: `account`,
  })
