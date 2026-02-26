import { BrowserRouter, Routes, Route, Outlet } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { ConfigProvider } from '@/providers/config-provider'
import { AuthProvider } from '@/providers/auth-provider'
import { OrgLayout } from '@/views/org/OrgLayout'
import { Login } from '@/views/Login'
import { Dashboard } from '@/views/org/Dashbaord'
import { Apps } from '@/views/org/Apps'
import { Installs } from '@/views/org/Installs'
import { BuildRunner } from '@/views/org/BuildRunner'
import { Team } from '@/views/org/Team'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})

const Onboarding = () => {
  return <>Onboarding</>
}

const AppLayout = () => {
  return (
    <div className="flex flex-col gap-4">
      <span>App layout</span>
      <Outlet />
    </div>
  )
}

export const App = () => {
  return (
    <ConfigProvider>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <BrowserRouter>
            <Routes>
              <Route index element={<Login />} />
              <Route element={<AuthLayout />}>
                <Route path="/onboarding" element={<Onboarding />} />
                <Route element={<OrgLayout />}>
                  <Route path=":orgId" element={<Dashboard />} />
                  <Route path=":orgId/apps" element={<Apps />} />
                  <Route path=":orgId/installs" element={<Installs />} />
                  <Route path=":orgId/runner" element={<BuildRunner />} />
                  <Route path=":orgId/team" element={<Team />} />

                  <Route element={<AppLayout />}>
                    <Route
                      path=":orgId/apps/:appId"
                      element={<>App overview</>}
                    />
                    <Route
                      path=":orgId/apps/:appId/components"
                      element={<>App components</>}
                    />
                  </Route>
                </Route>
              </Route>
            </Routes>
          </BrowserRouter>
        </AuthProvider>
        <ReactQueryDevtools />
      </QueryClientProvider>
    </ConfigProvider>
  )
}
