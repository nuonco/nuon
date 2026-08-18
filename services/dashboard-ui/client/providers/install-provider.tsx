import { createContext, useEffect, useMemo, type ReactNode } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getInstall, getAppLabels, toLabelColorMap } from '@/lib'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TInstall } from '@/types'

type InstallContextValue = {
  install: TInstall
  labelColors: Record<string, string>
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
    placeholderData: keepPreviousData,
    queryKey: ['install', org.id!, installId],
    queryFn: () => getInstall({ orgId: org.id!, installId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!installId,
  })

  const { data: labelsData } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-labels', org.id!, install?.app_id],
    queryFn: () => getAppLabels({ orgId: org.id!, appId: install!.app_id! }),
    enabled: !!org.id && !!install?.app_id,
  })

  const labelColors = useMemo(() => toLabelColorMap(labelsData), [labelsData])

  useEffect(() => {
    if (error && install) {
      addToast(
        <Toast heading="Refresh failed" theme="warn">
          <Text>{(error as TAPIError)?.error ?? 'Connection issue'}</Text>
        </Toast>
      )
    }
  }, [error])

  if (error && !install) return errorElement !== undefined ? <>{errorElement}</> : <ProviderError error={error} />

  if (isLoading || !install) return loadingElement

  return (
    <InstallContext.Provider value={{ install, labelColors, refresh: refetch }}>
      {children}
    </InstallContext.Provider>
  )
}
