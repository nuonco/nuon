import { useContext } from 'react'
import { InstallAppConfigContext } from '@/providers/install-app-config-provider'
import type { TAppConfig } from '@/types'

export function useInstallAppConfig(): {
  appConfig: TAppConfig | undefined
  isLoading: boolean
  error: unknown
  refresh: () => void
} {
  const ctx = useContext(InstallAppConfigContext)
  if (!ctx) {
    throw new Error(
      'useInstallAppConfig must be used within an InstallAppConfigProvider'
    )
  }
  return ctx
}
