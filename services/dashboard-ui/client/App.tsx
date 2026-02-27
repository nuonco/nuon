import { BrowserRouter, Routes, Route } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { ConfigProvider } from '@/providers/config-provider'
import { AuthProvider } from '@/providers/auth-provider'
import { Login } from '@/views/Login'
import { AppLayout } from '@/views/app/AppLayout'
import { Overview as AppOverview } from '@/views/app/Overview'
import { Components as AppComponents } from '@/views/app/Components'
import { Actions as AppActions } from '@/views/app/Actions'
import { Roles as AppRoles } from '@/views/app/Roles'
import { Policies as AppPolicies } from '@/views/app/Policies'
import { Installs as AppInstalls } from '@/views/app/Installs'
import { Readme as AppReadme } from '@/views/app/Readme'
import { InstallLayout } from '@/views/install/InstallLayout'
import { Overview as InstallOverview } from '@/views/install/Overview'
import { Components as InstallComponents } from '@/views/install/Components'
import { Actions as InstallActions } from '@/views/install/Actions'
import { Roles as InstallRoles } from '@/views/install/Roles'
import { Policies as InstallPolicies } from '@/views/install/Policies'
import { Runner as InstallRunner } from '@/views/install/Runner'
import { Sandbox as InstallSandbox } from '@/views/install/Sandbox'
import { Stacks as InstallStacks } from '@/views/install/Stacks'
import { Workflows as InstallWorkflows } from '@/views/install/Workflows'
import { Readme as InstallReadme } from '@/views/install/Readme'
import { OrgLayout } from '@/views/org/OrgLayout'
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
                      element={<AppOverview />}
                    />
                    <Route
                      path=":orgId/apps/:appId/components"
                      element={<AppComponents />}
                    />

                    <Route
                      path=":orgId/apps/:appId/actions"
                      element={<AppActions />}
                    />

                    <Route
                      path=":orgId/apps/:appId/roles"
                      element={<AppRoles />}
                    />

                    <Route
                      path=":orgId/apps/:appId/policies"
                      element={<AppPolicies />}
                    />

                    <Route
                      path=":orgId/apps/:appId/installs"
                      element={<AppInstalls />}
                    />

                    <Route
                      path=":orgId/apps/:appId/readme"
                      element={<AppReadme />}
                    />
                  </Route>

                  <Route element={<InstallLayout />}>
                    <Route
                      path=":orgId/installs/:installId"
                      element={<InstallOverview />}
                    />
                    <Route
                      path=":orgId/installs/:installId/components"
                      element={<InstallComponents />}
                    />
                    <Route
                      path=":orgId/installs/:installId/actions"
                      element={<InstallActions />}
                    />
                    <Route
                      path=":orgId/installs/:installId/roles"
                      element={<InstallRoles />}
                    />
                    <Route
                      path=":orgId/installs/:installId/policies"
                      element={<InstallPolicies />}
                    />
                    <Route
                      path=":orgId/installs/:installId/runner"
                      element={<InstallRunner />}
                    />
                    <Route
                      path=":orgId/installs/:installId/sandbox"
                      element={<InstallSandbox />}
                    />
                    <Route
                      path=":orgId/installs/:installId/stacks"
                      element={<InstallStacks />}
                    />
                    <Route
                      path=":orgId/installs/:installId/workflows"
                      element={<InstallWorkflows />}
                    />
                    <Route
                      path=":orgId/installs/:installId/readme"
                      element={<InstallReadme />}
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
