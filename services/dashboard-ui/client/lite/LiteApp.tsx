import { createBrowserRouter, RouterProvider } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from '@/providers/config-provider'
import type { TAPIError } from '@/types'
import { ThemeProvider } from './providers/theme-provider'
import { liteRoutes } from './routes'

const queryClient = new QueryClient({
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

const router = createBrowserRouter(liteRoutes)

export const LiteApp = () => (
  <ConfigProvider>
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </ThemeProvider>
  </ConfigProvider>
)
