import "../client/lite/styles.css"
import { useEffect } from "react"
import { ThemeState, type GlobalProvider } from "@ladle/react"
import { MemoryRouter } from "react-router"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ConfigContext, type TRuntimeConfig } from "@/providers/config-provider"
import {
  ThemeProvider,
  type TThemePreference,
} from "@/lite/providers/theme-provider"
import { useTheme } from "@/lite/hooks/use-theme"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false, staleTime: Infinity },
  },
})

const mockConfig: TRuntimeConfig = {
  apiUrl: "http://localhost:8081",
  appUrl: "http://localhost:4000",
  githubAppName: "nuon-dev",
  isByoc: false,
  dashboardLite: true,
}

const THEME_TO_PREFERENCE: Record<string, TThemePreference> = {
  [ThemeState.Light]: "light",
  [ThemeState.Dark]: "dark",
  [ThemeState.Auto]: "system",
}

const LadleThemeSync = ({ theme }: { theme: ThemeState }) => {
  const { setPreference } = useTheme()

  useEffect(() => {
    setPreference(THEME_TO_PREFERENCE[theme] ?? "system")
  }, [theme, setPreference])

  return null
}

export const Provider: GlobalProvider = ({ children, globalState }) => (
  <MemoryRouter>
    <QueryClientProvider client={queryClient}>
      <ConfigContext.Provider value={mockConfig}>
        <ThemeProvider>
          <LadleThemeSync theme={globalState.theme} />
          <div className="min-h-screen bg-surface-default text-primary">
            {children}
          </div>
        </ThemeProvider>
      </ConfigContext.Provider>
    </QueryClientProvider>
  </MemoryRouter>
)
