import "../client/styles.css"
import type { ReactNode } from "react"
import { GlobalProvider } from "@ladle/react"
import { MemoryRouter } from "react-router"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { AuthContext } from "@/providers/auth-provider"
import { ConfigContext, type TRuntimeConfig } from "@/providers/config-provider"
import { OrgContext } from "@/providers/org-provider"
import { InstallContext } from "@/providers/install-provider"
import { InstallAppConfigProvider } from "@/providers/install-app-config-provider"
import { SurfacesProvider } from "@/providers/surfaces-provider"
import { ToastProvider } from "@/providers/toast-provider"
import { DashboardPreferencesProvider } from "@/providers/dashboard-preferences-provider"
import { ThemeProvider } from "@/providers/theme-provider"
import { PageTitleProvider } from "@/providers/page-title-provider"
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
  runnerApiUrl: "http://localhost:8083",
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

const fullBleedStyles = `
  .ladle-main { padding: 0 !important; height: 100vh !important; overflow: hidden !important; }
`

const canvas = (isFullBleed: boolean, children: ReactNode) => (
  <>
    {/* The app themes via prefers-color-scheme, but Ladle's canvas
        follows its own toggle — paint the canvas with the app's
        background/foreground vars so stories stay readable when
        the OS is in dark mode. */}
    {isFullBleed && <style>{fullBleedStyles}</style>}
    <div
      className={
        isFullBleed
          ? "h-full bg-background text-foreground"
          : "min-h-screen bg-background text-foreground"
      }
    >
      {children}
    </div>
  </>
)

export const Provider: GlobalProvider = ({ children, storyMeta }) => {
  const meta = storyMeta as { fullBleed?: boolean; installViews?: boolean }
  const isFullBleed = Boolean(meta?.fullBleed)

  if (meta?.installViews) {
    return (
      <QueryClientProvider client={queryClient}>
        <ConfigContext.Provider value={mockConfig}>
          <AuthContext.Provider value={mockAuth}>
            <PageTitleProvider>
              <ToastProvider>
                <ThemeProvider>
                  <DashboardPreferencesProvider>
                    {canvas(true, children)}
                  </DashboardPreferencesProvider>
                </ThemeProvider>
              </ToastProvider>
            </PageTitleProvider>
          </AuthContext.Provider>
        </ConfigContext.Provider>
      </QueryClientProvider>
    )
  }

  return (
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>
        <ConfigContext.Provider value={mockConfig}>
          <AuthContext.Provider value={mockAuth}>
            <OrgContext.Provider value={{ org: mockOrg, refresh: () => {} }}>
              <InstallContext.Provider value={{ install: mockInstall, refresh: () => {} }}>
                <InstallAppConfigProvider>
                  <ToastProvider>
                    <ThemeProvider>
                      <DashboardPreferencesProvider>
                        <SurfacesProvider>
                          {canvas(isFullBleed, children)}
                        </SurfacesProvider>
                      </DashboardPreferencesProvider>
                    </ThemeProvider>
                  </ToastProvider>
                </InstallAppConfigProvider>
              </InstallContext.Provider>
            </OrgContext.Provider>
          </AuthContext.Provider>
        </ConfigContext.Provider>
      </QueryClientProvider>
    </MemoryRouter>
  )
}
