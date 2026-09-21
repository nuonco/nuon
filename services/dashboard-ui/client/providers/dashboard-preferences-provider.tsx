import { createContext, useMemo, useState } from 'react'
import { useConfig } from '@/hooks/use-config'
import {
  getInstallsTabEnabled,
  getStatusBarEnabled,
  setInstallsTabEnabled,
  setStatusBarEnabled,
} from '@/lib/cookies'

export interface IDashboardPreferencesContext {
  isStatusBarEnabled: boolean
  setIsStatusBarEnabled: (isEnabled: boolean) => void
  isInstallsTabEnabled: boolean
  setIsInstallsTabEnabled: (isEnabled: boolean) => void
}

export const DashboardPreferencesContext = createContext<
  IDashboardPreferencesContext | undefined
>(undefined)

export const DashboardPreferencesProvider = ({
  children,
}: {
  children: React.ReactNode
}) => {
  const { installsTabAutoEnabled, statusBarAutoEnabled } = useConfig()
  const [isStatusBarEnabled, setIsStatusBarEnabledState] = useState(
    () => getStatusBarEnabled() ?? statusBarAutoEnabled ?? true
  )
  const [isInstallsTabEnabled, setIsInstallsTabEnabledState] = useState(
    () => getInstallsTabEnabled() ?? installsTabAutoEnabled ?? true
  )

  const value = useMemo(
    () => ({
      isStatusBarEnabled,
      setIsStatusBarEnabled: (isEnabled: boolean) => {
        setStatusBarEnabled(isEnabled)
        setIsStatusBarEnabledState(isEnabled)
      },
      isInstallsTabEnabled,
      setIsInstallsTabEnabled: (isEnabled: boolean) => {
        setInstallsTabEnabled(isEnabled)
        setIsInstallsTabEnabledState(isEnabled)
      },
    }),
    [isInstallsTabEnabled, isStatusBarEnabled]
  )

  return (
    <DashboardPreferencesContext.Provider value={value}>
      {children}
    </DashboardPreferencesContext.Provider>
  )
}
