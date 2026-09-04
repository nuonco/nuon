import { createBrowserRouter, RouterProvider } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { WorkerPoolContextProvider } from '@pierre/diffs/react'
import { ConfigProvider } from '@/providers/config-provider'
import type { TAPIError } from '@/types'
import { LITE_SYNTAX_THEME, registerSyntax, SUPPORTED_LANGUAGES } from './lib/syntax'
import { workerFactory } from './lib/syntax/worker-pool'
import { ThemeProvider } from './providers/theme-provider'
import { liteRoutes } from './routes'

registerSyntax()

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
        <WorkerPoolContextProvider
          poolOptions={{ workerFactory }}
          highlighterOptions={{
            theme: LITE_SYNTAX_THEME,
            langs: [...SUPPORTED_LANGUAGES],
          }}
        >
          <RouterProvider router={router} />
        </WorkerPoolContextProvider>
      </QueryClientProvider>
    </ThemeProvider>
  </ConfigProvider>
)
