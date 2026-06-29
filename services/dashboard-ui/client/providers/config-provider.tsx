import { createContext, useMemo } from 'react'

export type TRuntimeConfig = {
  apiUrl: string

  temporalUiUrl?: string
  authServiceUrl?: string
  appUrl: string
  githubAppName: string
  pylonAppId?: string
  datadogEnv?: string
  datadogApiKey?: string
  datadogApplicationKey?: string
  datadogTraceDebug?: boolean
  datadogApiUrl?: string
  version?: string
  gitRef?: string
  isByoc: boolean
  onboardingV2?: boolean
  adminDashboardUrl?: string
  isDev?: boolean
}

declare global {
  interface Window {
    __NUON_CONFIG__?: TRuntimeConfig
  }
}

export const ConfigContext = createContext<TRuntimeConfig | undefined>(
  undefined
)

export const ConfigProvider = ({ children }: { children: React.ReactNode }) => {
  const config = useMemo<TRuntimeConfig>(() => {
    const cfg = window.__NUON_CONFIG__ ?? ({} as TRuntimeConfig)
    document.getElementById('nuon-config')?.remove()
    delete window.__NUON_CONFIG__
    const apiUrl = cfg.apiUrl ?? ''
    const isDev =
      apiUrl.includes('localhost') || apiUrl.includes('127.0.0.1')
    return { ...cfg, isDev }
  }, [])

  return (
    <ConfigContext.Provider value={config}>{children}</ConfigContext.Provider>
  )
}
