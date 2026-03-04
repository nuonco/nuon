import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallSandboxRun } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TSandboxRun } from '@/types'

type SandboxRunContextValue = {
  sandboxRun: TSandboxRun
}

export const SandboxRunContext = createContext<SandboxRunContextValue | undefined>(
  undefined
)

export function SandboxRunProvider({
  children,
  runId,
  pollInterval = 10000,
  shouldPoll = true,
}: {
  children: ReactNode
  runId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { data: sandboxRun, isLoading, error } = useQuery({
    queryKey: ['sandbox-run', org.id!, runId],
    queryFn: () => getInstallSandboxRun({ orgId: org.id!, runId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!runId,
  })

  if (error) return <ProviderError error={error} />

  if (isLoading || !sandboxRun) return <ProviderLoading />

  return (
    <SandboxRunContext.Provider value={{ sandboxRun }}>
      {children}
    </SandboxRunContext.Provider>
  )
}
