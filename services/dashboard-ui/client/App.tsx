import { createBrowserRouter, RouterProvider } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { APIHealthProvider } from '@/providers/api-health-provider'
import { AuthProvider } from '@/providers/auth-provider'
import { ConfigProvider } from '@/providers/config-provider'
import { Login } from '@/views/Login'
import { NotFound } from '@/views/NotFound'
import { RouteError } from '@/views/RouteError'
import { Onboarding } from '@/views/Onboarding'
import { orgRoutes } from '@/views/org/routes'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

const router = createBrowserRouter([
  { index: true, element: <Login /> },
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
