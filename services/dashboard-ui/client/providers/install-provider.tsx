import { createContext, useEffect, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getInstall } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TInstall } from '@/types'

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
  isSkeletonLoading = false,
  loadingElement = <ProviderLoading />,
  errorElement,
}: {
  children: ReactNode
  installId: string
  pollInterval?: number
  shouldPoll?: boolean
  isSkeletonLoading?: boolean
  loadingElement?: ReactNode
  errorElement?: ReactNode
}) {
  const { org } = useOrg()
  const { addToast } = useToast()
  const {
    data: install,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['install', org.id!, installId],
    queryFn: () => getInstall({ orgId: org.id!, installId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!installId,
  })

  useEffect(() => {
    if (error && install) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !install) return errorElement !== undefined ? <>{errorElement}</> : <ProviderError error={error} />

  if (isLoading || !install) return loadingElement

  return (
    <InstallContext.Provider value={{ install, refresh: refetch }}>
      {children}
    </InstallContext.Provider>
  )
}
