import { createContext, useContext, useMemo } from 'react'

export type TRuntimeConfig = {
  apiUrl: string
  adminApiUrl?: string
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
  sfTrialEndpoint?: string
}

declare global {
  interface Window {
    __NUON_CONFIG__?: TRuntimeConfig
  }
}

export const ConfigContext = createContext<TRuntimeConfig | undefined>(undefined)

export const ConfigProvider = ({ children }: { children: React.ReactNode }) => {
  const config = useMemo(() => {
    const cfg = window.__NUON_CONFIG__ ?? ({} as TRuntimeConfig)
    document.getElementById('nuon-config')?.remove()
    delete window.__NUON_CONFIG__
    return cfg
  }, [])

  return <ConfigContext.Provider value={config}>{children}</ConfigContext.Provider>
}
