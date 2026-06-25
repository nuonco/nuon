import { createContext, useEffect, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useApp } from '@/hooks/use-app'
import { useToast } from '@/hooks/use-toast'
import { getAppBranch } from '@/lib'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TAppBranch } from '@/types'

type BranchContextValue = {
  branch: TAppBranch
  refresh: () => void
}

export const BranchContext = createContext<BranchContextValue | undefined>(undefined)

export function BranchProvider({
  children,
  branchId,
  pollInterval = 20000,
  shouldPoll = false,
  loadingElement = <ProviderLoading />,
  errorElement,
}: {
  children: ReactNode
  branchId: string
  pollInterval?: number
  shouldPoll?: boolean
  loadingElement?: ReactNode
  errorElement?: ReactNode
}) {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const { data: branch, isLoading, error, refetch } = useQuery({
    queryKey: ['app-branch', org.id!, app.id!, branchId, 'with-config'],
    queryFn: () => getAppBranch({ orgId: org.id!, appId: app.id!, branchId, latestConfig: true }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!app.id && !!branchId,
  })

  useEffect(() => {
    if (error && branch) {
      addToast(
        <Toast heading="Refresh failed" theme="warn">
          <Text>{(error as TAPIError)?.error ?? 'Connection issue'}</Text>
        </Toast>
      )
    }
  }, [error])

  if (error && !branch) return errorElement !== undefined ? <>{errorElement}</> : <ProviderError error={error} />

  if (isLoading || !branch) return <>{loadingElement}</>

  return (
    <BranchContext.Provider value={{ branch, refresh: refetch }}>
      {children}
    </BranchContext.Provider>
  )
}
