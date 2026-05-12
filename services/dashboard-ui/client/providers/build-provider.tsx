import { createContext, useEffect, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useStatusToast } from '@/hooks/use-status-toast'
import { useToast } from '@/hooks/use-toast'
import { getComponentBuild } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TBuild } from '@/types'

type BuildContextValue = {
  build: TBuild
}

export const BuildContext = createContext<BuildContextValue | undefined>(
  undefined
)

export function BuildProvider({
  children,
  buildId,
  componentId,
  componentName,
  pollInterval = 10000,
  shouldPoll = true,
}: {
  children: ReactNode
  buildId: string
  componentId: string
  componentName?: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { addToast } = useToast()
  const { data: build, isLoading, error } = useQuery({
    queryKey: ['build', org.id!, componentId, buildId],
    queryFn: () => getComponentBuild({ orgId: org.id!, componentId, buildId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!componentId && !!buildId,
  })

  useStatusToast({
    status: build?.status_v2?.status,
    label: componentName ?? build?.component_name,
    resourceType: 'build',
  })

  useEffect(() => {
    if (error && build) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !build) return <ProviderError error={error} />

  if (isLoading || !build) return <ProviderLoading />

  return (
    <BuildContext.Provider value={{ build }}>
      {children}
    </BuildContext.Provider>
  )
}
