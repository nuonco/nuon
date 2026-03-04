import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
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
  pollInterval = 10000,
  shouldPoll = true,
}: {
  children: ReactNode
  buildId: string
  componentId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { data: build, isLoading, error } = useQuery({
    queryKey: ['build', org.id!, componentId, buildId],
    queryFn: () => getComponentBuild({ orgId: org.id!, componentId, buildId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!componentId && !!buildId,
  })

  if (error) return <ProviderError error={error} />

  if (isLoading || !build) return <ProviderLoading />

  return (
    <BuildContext.Provider value={{ build }}>
      {children}
    </BuildContext.Provider>
  )
}
