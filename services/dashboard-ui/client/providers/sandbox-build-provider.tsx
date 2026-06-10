import { createContext, type ReactNode } from 'react'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSSEResourceQuery, isTerminalStatusV2 } from '@/hooks/use-sse-resource-query'
import { useStatusToast } from '@/hooks/use-status-toast'
import { getSandboxBuild } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAppSandboxBuild } from '@/types'

type SandboxBuildContextValue = {
  build: TAppSandboxBuild
}

export const SandboxBuildContext = createContext<
  SandboxBuildContextValue | undefined
>(undefined)

export function SandboxBuildProvider({
  children,
  buildId,
  shouldPoll = true,
}: {
  children: ReactNode
  buildId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { app } = useApp()

  const { data: build, isLoading, error } = useSSEResourceQuery<TAppSandboxBuild>({
    sseUrl: org?.id && app?.id && buildId
      ? `/api/orgs/${org.id}/apps/${app.id}/sandbox-builds/${buildId}/sse`
      : undefined,
    queryKey: ['sandbox-build', org?.id, app?.id, buildId],
    queryFn: () => getSandboxBuild({ orgId: org!.id, appId: app!.id, buildId }),
    enabled: !!org?.id && !!app?.id && !!buildId,
    shouldPoll,
    eventName: 'sandbox-build',
    isFinished: isTerminalStatusV2,
  })

  useStatusToast({
    status: build?.status_v2?.status,
    resourceType: 'sandbox build',
  })

  if (error && !build) return <ProviderError error={error} />
  if (isLoading || !build) return <ProviderLoading />

  return (
    <SandboxBuildContext.Provider value={{ build }}>
      {children}
    </SandboxBuildContext.Provider>
  )
}
