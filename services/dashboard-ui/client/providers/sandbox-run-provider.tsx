import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallSandboxRun } from '@/lib'
import { Loading } from '@/components/common/Loading'
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
  const { data: sandboxRun, isLoading } = useQuery({
    queryKey: ['sandbox-run', org.id!, runId],
    queryFn: () => getInstallSandboxRun({ orgId: org.id!, runId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!runId,
  })

  if (isLoading || !sandboxRun) return <Loading />

  return (
    <SandboxRunContext.Provider value={{ sandboxRun }}>
      {children}
    </SandboxRunContext.Provider>
  )
}
