import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstall } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TInstall } from '@/types'

type InstallContextValue = {
  install: TInstall
  refresh: () => void
}

export const InstallContext = createContext<InstallContextValue | undefined>(
  undefined
)

export function InstallProvider({
  children,
  installId,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  installId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { data: install, isLoading, error, refetch } = useQuery({
    queryKey: ['install', org.id!, installId],
    queryFn: () => getInstall({ orgId: org.id!, installId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!installId,
  })

  if (error) return <ProviderError error={error} />

  if (isLoading || !install) return <ProviderLoading />

  return (
    <InstallContext.Provider value={{ install, refresh: refetch }}>
      {children}
    </InstallContext.Provider>
  )
}
