import { QueryClient } from '@tanstack/react-query'
import type { TAPIError } from '@/types'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: 'always',
      retry: (failureCount, error) => {
        const status = (error as TAPIError)?.status
        if (status && status >= 400 && status < 500) return false
        return failureCount < 3
      },
    },
  },
})
