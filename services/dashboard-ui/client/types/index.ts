export * from '@/types/ctl-api.types'
export * from '@/types/dashboard.types'

import type { TAPIError } from '@/types/dashboard.types'

declare module '@tanstack/react-query' {
  interface Register {
    defaultError: TAPIError
  }
}
