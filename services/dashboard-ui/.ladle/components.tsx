import "../client/styles.css"
import { GlobalProvider } from "@ladle/react"
import { MemoryRouter } from "react-router"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { AuthContext } from "@/providers/auth-provider"
import { ConfigContext, type TRuntimeConfig } from "@/providers/config-provider"
import { OrgContext } from "@/providers/org-provider"
import { InstallContext } from "@/providers/install-provider"
import { SurfacesProvider } from "@/providers/surfaces-provider"
import { ToastProvider } from "@/providers/toast-provider"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false, staleTime: Infinity },
  },
})

const mockUser = {
  sub: "user-001",
  email: "jane@example.com",
  name: "Jane Doe",
  picture: undefined,
}

const mockAuth = {
  user: mockUser,
  isAuthenticated: true,
  isAdmin: false,
  isLoading: false,
  error: null,
}

const mockConfig: TRuntimeConfig = {
  apiUrl: "http://localhost:8081",
  appUrl: "http://localhost:4000",
  githubAppName: "nuon-dev",
  isByoc: false,
}

const mockOrg = {
  id: "org-mock-001",
  name: "Mock Organization",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
} as any

const mockInstall = {
  id: "inst-mock-001",
  name: "mock-install",
  org_id: "org-mock-001",
  app_id: "app-mock-001",
  runner_status: "active",
  sandbox_status: "active",
  composite_component_status: "active",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
} as any

export const Provider: GlobalProvider = ({ children }) => {
  return (
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>
        <ConfigContext.Provider value={mockConfig}>
          <AuthContext.Provider value={mockAuth}>
            <OrgContext.Provider value={{ org: mockOrg, refresh: () => {} }}>
              <InstallContext.Provider value={{ install: mockInstall, refresh: () => {} }}>
                <ToastProvider>
                  <SurfacesProvider>
                    {children}
                  </SurfacesProvider>
                </ToastProvider>
              </InstallContext.Provider>
            </OrgContext.Provider>
          </AuthContext.Provider>
        </ConfigContext.Provider>
      </QueryClientProvider>
    </MemoryRouter>
  )
}
