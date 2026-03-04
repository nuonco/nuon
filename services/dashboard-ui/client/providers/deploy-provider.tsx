import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getDeploy } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TDeploy } from '@/types'

type DeployContextValue = {
  deploy: TDeploy
}

export const DeployContext = createContext<DeployContextValue | undefined>(
  undefined
)

export function DeployProvider({
  children,
  deployId,
  installId,
  pollInterval = 10000,
  shouldPoll = true,
}: {
  children: ReactNode
  deployId: string
  installId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { data: deploy, isLoading, error } = useQuery({
    queryKey: ['deploy', org.id!, installId, deployId],
    queryFn: () => getDeploy({ orgId: org.id!, installId, deployId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!installId && !!deployId,
  })

  if (error) return <ProviderError error={error} />

  if (isLoading || !deploy) return <ProviderLoading />

  return (
    <DeployContext.Provider value={{ deploy }}>
      {children}
    </DeployContext.Provider>
  )
}
