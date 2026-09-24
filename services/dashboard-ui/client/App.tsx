import { createBrowserRouter, RouterProvider } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { TAPIError } from '@/types'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { WorkerPoolContextProvider } from '@pierre/diffs/react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { APIHealthProvider } from '@/providers/api-health-provider'
import { AuthProvider } from '@/providers/auth-provider'
import { ConfigProvider } from '@/providers/config-provider'
import { PageTitleProvider } from '@/providers/page-title-provider'
import { ThemeProvider } from '@/providers/theme-provider'
import { SYNTAX_THEME, registerSyntax, SUPPORTED_LANGUAGES } from '@/lib/syntax'
import { workerFactory } from '@/lib/syntax/worker-pool'
import { Error } from '@/views/Error'
import { NotFound } from '@/views/NotFound'
import { RouteError } from '@/views/RouteError'
import { Onboarding } from '@/views/Onboarding'
import { BYOCSetup } from '@/views/BYOCSetup'
import { orgRoutes } from '@/views/org/routes'

registerSyntax()

const BFFRedirect = () => {
  const search = new URLSearchParams(window.location.search)
  if (search.get('slack') === 'installed') {
    const orgId = search.get('org_id')
    if (orgId) {
      window.location.replace(`/${orgId}/slack?slack=installed`)
      return null
    }
  }
  window.location.href = '/'
  return null
}

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

const router = createBrowserRouter([
  { index: true, element: <BFFRedirect /> },
  {
    element: <AuthLayout />,
    errorElement: <RouteError />,
    children: [
      { path: '/error', element: <Error /> },
      { path: '/onboarding', element: <Onboarding /> },
      { path: '/byoc-setup', element: <BYOCSetup /> },
      ...orgRoutes,
      { path: '*', element: <NotFound /> },
    ],
  },
])

export const App = () => {
  return (
    <ThemeProvider>
      <ConfigProvider>
        <QueryClientProvider client={queryClient}>
          <AuthProvider>
            <APIHealthProvider shouldPoll>
              <PageTitleProvider>
                <WorkerPoolContextProvider
                  poolOptions={{ workerFactory }}
                  highlighterOptions={{
                    theme: SYNTAX_THEME,
                    langs: [...SUPPORTED_LANGUAGES],
                  }}
                >
                  <RouterProvider router={router} />
                </WorkerPoolContextProvider>
              </PageTitleProvider>
            </APIHealthProvider>
          </AuthProvider>
          <ReactQueryDevtools />
        </QueryClientProvider>
      </ConfigProvider>
    </ThemeProvider>
  )
}
