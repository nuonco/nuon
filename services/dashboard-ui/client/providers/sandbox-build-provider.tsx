import { createContext, useEffect, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getSandboxBuild } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TAppSandboxBuild } from '@/types'

type SandboxBuildContextValue = {
  build: TAppSandboxBuild
}

export const SandboxBuildContext = createContext<
  SandboxBuildContextValue | undefined
>(undefined)

export function SandboxBuildProvider({
  children,
  buildId,
  pollInterval = 10000,
  shouldPoll = true,
}: {
  children: ReactNode
  buildId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()

  const {
    data: build,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['sandbox-build', org.id, app.id, buildId],
    queryFn: () => getSandboxBuild({ orgId: org.id, appId: app.id, buildId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!app.id && !!buildId,
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
    <SandboxBuildContext.Provider value={{ build }}>
      {children}
    </SandboxBuildContext.Provider>
  )
}
