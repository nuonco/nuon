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
  initInstall,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  initInstall: TInstall
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { data: install = initInstall, error, isLoading, refetch } = useQuery({
    queryKey: ['install', org.id, initInstall.id],
    queryFn: () => getInstall({ orgId: org.id, installId: initInstall.id }),
    initialData: initInstall,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return (
    <InstallContext.Provider
      value={{
        install,
        isLoading,
        error,
        refresh: refetch,
      }}
    >
      {children}
    </InstallContext.Provider>
  )
}
