import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallSandboxRun } from '@/lib'
import type { TSandboxRun } from '@/types'

type SandboxRunContextValue = {
  sandboxRun: TSandboxRun | null
  isLoading: boolean
  error: any
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
  const {
    data: sandboxRun,
    error,
    isLoading,
  } = useQuery({
    queryKey: ['sandbox-run', org?.id, runId],
    queryFn: () => getInstallSandboxRun({ orgId: org.id, runId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runId,
  })

  return (
    <SandboxRunContext.Provider
      value={{
        sandboxRun: sandboxRun ?? null,
        isLoading,
        error,
      }}
    >
      {children}
    </SandboxRunContext.Provider>
  )
}
