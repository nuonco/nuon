import { createBrowserRouter, RouterProvider } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { TAPIError } from '@/types'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { APIHealthProvider } from '@/providers/api-health-provider'
import { AuthProvider } from '@/providers/auth-provider'
import { ConfigProvider } from '@/providers/config-provider'
import { NotFound } from '@/views/NotFound'
import { RouteError } from '@/views/RouteError'
import { Onboarding } from '@/views/Onboarding'
import { orgRoutes } from '@/views/org/routes'

const BFFRedirect = () => {
  window.location.href = '/'
  return null
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
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
      { path: '/onboarding', element: <Onboarding /> },
      ...orgRoutes,
      { path: '*', element: <NotFound /> },
    ],
  },
])

export const App = () => {
  return (
    <ConfigProvider>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <APIHealthProvider shouldPoll>
            <RouterProvider router={router} />
          </APIHealthProvider>
        </AuthProvider>
        <ReactQueryDevtools />
      </QueryClientProvider>
    </ConfigProvider>
  )
}
