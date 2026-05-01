import { createContext, useEffect, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getInstallSandboxRun } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TSandboxRun } from '@/types'

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
  const { addToast } = useToast()
  const { data: sandboxRun, isLoading, error } = useQuery({
    queryKey: ['sandbox-run', org.id!, runId],
    queryFn: () => getInstallSandboxRun({ orgId: org.id!, runId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!runId,
  })

  useEffect(() => {
    if (error && sandboxRun) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !sandboxRun) return <ProviderError error={error} />

  if (isLoading || !sandboxRun) return <ProviderLoading />

  return (
    <SandboxRunContext.Provider value={{ sandboxRun }}>
      {children}
    </SandboxRunContext.Provider>
  )
}
