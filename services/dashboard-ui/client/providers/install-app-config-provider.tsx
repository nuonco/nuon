import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { getAppConfig } from '@/lib'
import type { TAppConfig } from '@/types'

type InstallAppConfigContextValue = {
  appConfig: TAppConfig | undefined
  isLoading: boolean
  error: unknown
  refresh: () => void
}

export const InstallAppConfigContext = createContext<
  InstallAppConfigContextValue | undefined
>(undefined)

export function InstallAppConfigProvider({
  children,
  enabled = true,
}: {
  children: ReactNode
  enabled?: boolean
}) {
  const { org } = useOrg()
  const { install } = useInstall()

  const {
    data: appConfig,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: [
      'install-app-config',
      org?.id,
      install?.app_id,
      install?.app_config_id,
      'recurse',
    ],
    queryFn: () =>
      getAppConfig({
        orgId: org.id!,
        appId: install.app_id!,
        appConfigId: install.app_config_id!,
        recurse: true,
      }),
    enabled: enabled && !!org?.id && !!install?.app_id && !!install?.app_config_id,
  })

  return (
    <InstallAppConfigContext.Provider
      value={{ appConfig, isLoading, error, refresh: refetch }}
    >
      {children}
    </InstallAppConfigContext.Provider>
  )
}
