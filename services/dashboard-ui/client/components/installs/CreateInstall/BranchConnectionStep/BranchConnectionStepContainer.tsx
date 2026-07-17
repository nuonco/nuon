import { useEffect } from 'react'
import { useQuery, useQueries } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getAppBranches, getAppBranch } from '@/lib'
import { BranchConnectionStep } from './BranchConnectionStep'
import { Text } from '@/components/common/Text'

interface IBranchConnectionStepContainer {
  appId: string
  installId: string
  onDone: () => void
  onSkip: () => void
}

export const BranchConnectionStepContainer = ({
  appId,
  installId,
  onDone,
  onSkip,
}: IBranchConnectionStepContainer) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''

  const { data: branchList, isLoading, isSuccess } = useQuery({
    queryKey: ['app-branches', orgId, appId],
    queryFn: () => getAppBranches({ appId, orgId: orgId! }),
    enabled: !!orgId && !!appId,
  })

  const branchIds = (branchList?.data ?? []).map((b) => b?.id).filter(Boolean) as string[]
  const hasBranches = branchIds.length > 0

  const branchQueries = useQueries({
    queries: hasBranches
      ? branchIds.map((branchId) => ({
          queryKey: ['app-branch-with-config', orgId, appId, branchId],
          queryFn: () => getAppBranch({ appId, branchId, orgId: orgId!, latestConfig: true }),
          enabled: !!orgId && !!appId,
        }))
      : [],
  })

  const branchesWithConfigs = branchQueries
    .map((q) => q.data)
    .filter(Boolean)

  useEffect(() => {
    if (isSuccess && !hasBranches) {
      onDone()
    }
  }, [isSuccess, hasBranches, onDone])

  if (isLoading) {
    return (
      <Text variant="body" theme="neutral">
        Loading app branches...
      </Text>
    )
  }

  if (!hasBranches) return null

  return (
    <BranchConnectionStep
      branches={branchesWithConfigs as any[]}
      installId={installId}
      orgId={orgId}
      appId={appId}
      onDone={onDone}
      onSkip={onSkip}
    />
  )
}
