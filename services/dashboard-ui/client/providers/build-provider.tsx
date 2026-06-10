import { createContext, type ReactNode } from 'react'
import { useOrg } from '@/hooks/use-org'
import { useSSEResourceQuery, isTerminalStatusV2 } from '@/hooks/use-sse-resource-query'
import { useStatusToast } from '@/hooks/use-status-toast'
import { getComponentBuild } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TBuild } from '@/types'

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

  const { data: build, isLoading, error } = useSSEResourceQuery<TBuild>({
    sseUrl: org?.id && componentId && buildId
      ? `/api/orgs/${org.id}/components/${componentId}/builds/${buildId}/sse`
      : undefined,
    queryKey: ['build', org?.id, componentId, buildId],
    queryFn: () => getComponentBuild({ orgId: org!.id, componentId, buildId }),
    enabled: !!org?.id && !!componentId && !!buildId,
    shouldPoll,
    eventName: 'build',
    isFinished: isTerminalStatusV2,
  })

  useStatusToast({
    status: build?.status_v2?.status,
    label: componentName ?? build?.component_name,
    resourceType: 'build',
  })

  if (error && !build) return <ProviderError error={error} />
  if (isLoading || !build) return <ProviderLoading />

  return (
    <BuildContext.Provider value={{ build }}>
      {children}
    </BuildContext.Provider>
  )
}
