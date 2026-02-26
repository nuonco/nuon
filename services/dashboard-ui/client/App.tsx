import { BrowserRouter, Routes, Route, Outlet } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { ConfigProvider } from '@/providers/config-provider'
import { AuthProvider } from '@/providers/auth-provider'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { Login } from '@/views/Login'

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

const OrgLayout = () => {
  return (
    <div className="flex flex-col gap-8 p-12">
      <span>Org layout</span>
      <Outlet />
    </div>
  )
}

const OrgApps = () => {
  return <>Org apps</>
}
const OrgInstalls = () => {
  return <>Org installs</>
}
const OrgDashboard = () => {
  return <>Org dashboard</>
}
const OrgRunner = () => {
  return <>Org runner</>
}
const OrgTeam = () => {
  return <>Org team</>
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
                  <Route path=":orgId" element={<OrgDashboard />} />
                  <Route path=":orgId/apps" element={<OrgApps />} />
                  <Route path=":orgId/installs" element={<OrgInstalls />} />
                  <Route path=":orgId/runner" element={<OrgRunner />} />
                  <Route path=":orgId/team" element={<OrgTeam />} />
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
