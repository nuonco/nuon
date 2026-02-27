import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstall } from '@/lib'
import type { TInstall } from '@/types'

type InstallContextValue = {
  install: TInstall | null
  isLoading: boolean
  error: any
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
  const { data: install, error, isLoading, refetch } = useQuery({
    queryKey: ['install', org?.id, installId],
    queryFn: () => getInstall({ orgId: org.id, installId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!installId,
  })

  return (
    <InstallContext.Provider
      value={{
        install: install ?? null,
        isLoading,
        error,
        refresh: refetch,
      }}
    >
      {children}
    </InstallContext.Provider>
  )
}
