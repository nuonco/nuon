import { createContext, useContext, useMemo, type ReactNode } from 'react'
import { enabledModules, type TModuleId } from '../utils/modules'
import { useOrg } from './org-provider'

export interface IModulesContext {
  enabled: ReadonlySet<TModuleId>
  has: (id: TModuleId) => boolean
  ready: boolean
}

export const ModulesContext = createContext<IModulesContext | null>(null)

export const ModulesProvider = ({ children }: { children: ReactNode }) => {
  const { org, loading } = useOrg()

  const value = useMemo<IModulesContext>(() => {
    const enabled = enabledModules(org?.features)
    return {
      enabled,
      has: (id) => enabled.has(id),
      ready: !!org || !loading,
    }
  }, [loading, org])

  return (
    <ModulesContext.Provider value={value}>{children}</ModulesContext.Provider>
  )
}

export const useModules = () => {
  const context = useContext(ModulesContext)
  if (!context) {
    throw new Error('useModules must be used within ModulesProvider')
  }
  return context
}
