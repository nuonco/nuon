import React, { createContext, useContext, useMemo } from 'react'
import { setBasePath } from '@/lib/api'

export type TAdminConfig = {
  appUrl: string
  basePath: string
}

const ConfigContext = createContext<TAdminConfig>({} as TAdminConfig)

export const useConfig = () => useContext(ConfigContext)

declare global {
  interface Window {
    __ADMIN_CONFIG__?: TAdminConfig
  }
}

export const ConfigProvider = ({ children }: { children: React.ReactNode }) => {
  const config = useMemo<TAdminConfig>(() => {
    const cfg = window.__ADMIN_CONFIG__ ?? ({} as TAdminConfig)
    cfg.basePath = cfg.basePath || ''
    setBasePath(cfg.basePath)
    document.getElementById('admin-config')?.remove()
    delete window.__ADMIN_CONFIG__
    return cfg
  }, [])

  return <ConfigContext.Provider value={config}>{children}</ConfigContext.Provider>
}
